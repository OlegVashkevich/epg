package main

import (
	"bufio"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Channel struct {
	ID          string `xml:"id,attr"`
	DisplayName string `xml:"display-name"`
}

type Programme struct {
	Start   string `xml:"start,attr"`
	Stop    string `xml:"stop,attr"`
	Channel string `xml:"channel,attr"`
	Title   string `xml:"title"`
}

type epgData struct {
	XMLName    xml.Name    `xml:"tv"`
	Channels   []Channel   `xml:"channel"`
	Programmes []Programme `xml:"programme"`
}

func main() {
	var text string
	var e error

	epgData, err := prepare()
	if err != nil {
		fmt.Println("Ошибка при подготовке данных:", err)
		return
	}

	// Бесконечный цикл для постоянного поиска
	for e == nil {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Введите слово или точную фразу для поиска: ")
		text, _ = reader.ReadString('\n')
		// Удаляем символы переноса строки
		text = strings.Replace(text, "\r\n", "", -1)
		searchWord := strings.ToLower(text)

		// Создаем словарь для хранения ID и отображений каналов
		channels := make(map[string]string)
		for _, channel := range epgData.Channels {
			channels[channel.ID] = channel.DisplayName
		}

		var founds = make(map[string][]string)

		// Проходим по всем программам и проверяем совпадение с ключевым словом
		layout := "20060102150405"
		for _, programme := range epgData.Programmes {
			title := strings.ToLower(programme.Title)
			if strings.Contains(title, searchWord) {
				start := programme.Start[:14]
				end := programme.Stop[:14]
				startTime, err := time.Parse(layout, start)
				if err != nil {
					fmt.Println("Error parsing start time:", err)
					continue
				}
				endTime, err := time.Parse(layout, end)
				if err != nil {
					fmt.Println("Error parsing end time:", err)
					continue
				}
				// Форматируем даты для удобства вывода
				startFormatted := startTime.Format("02.01.2006 15:04")
				endFormatted := endTime.Format("15:04")
				// Получаем название канала по его ID
				displayName := channels[programme.Channel]
				founds[displayName] = append(founds[displayName], fmt.Sprintf("%s - %s %s", startFormatted, endFormatted, programme.Title))
			}
		}

		// Выводим результаты поиска
		for key, value := range founds {
			fmt.Println(key)
			for _, item := range value {
				fmt.Println("  ", item)
			}
		}
	}
}

// Функция для проверки существования файла
func fileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

// Функция для получения последнего изменения файла
func fileLastModified(filename string) int64 {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return 0
	}
	return fileInfo.ModTime().Unix()
}

// Функция для скачивания файла
func downloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func prepare() (epgData, error) {

	epgData := new(epgData)

	filePath := "epg2.xml.gz"

	// Проверяем наличие файла или его актуальность
	if !fileExists(filePath) || fileLastModified(filePath) < time.Now().Add(-24*time.Hour).Unix() {
		url := "http://epg.one/epg2.xml.gz"
		// Загружаем файл с программой телевещания
		downloadFile(url, filePath)
	}

	xmlReader, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return *epgData, err
	}
	defer xmlReader.Close()

	gzReader, err := gzip.NewReader(xmlReader)
	if err != nil {
		fmt.Println("Error creating gzip reader:", err)
		return *epgData, err
	}
	defer gzReader.Close()

	decoder := xml.NewDecoder(gzReader)
	fmt.Println("Обрабатываю данные, подождите")
	err = decoder.Decode(&epgData)
	if err != nil {
		fmt.Println("Error decoding XML:", err)
		return *epgData, err
	}
	return *epgData, nil
}
