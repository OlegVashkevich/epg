<?php

if (isset($argv[1])) {
    $file = "epg2.xml.gz";
    // Check file date
    if (!file_exists($file) || filemtime("epg2.xml.gz") < time() - 86400) {
        // Load file from web
        $url = "http://epg.one/epg2.xml.gz";
        file_put_contents($file, file_get_contents($url));
    }
    $fraze = mb_strtolower($argv[1]);

    $reader = new XMLReader();
    $reader->open("compress.zlib://{$file}");

    $channels = [];
    $founds = [];

    while ($reader->read()) {
        if ($reader->nodeType == XMLReader::ELEMENT) {
            switch ($reader->name) {
                case 'channel':
                    $channel_id = (int)$reader->getAttribute('id');
                    while ($reader->read()) {
                        if ($reader->nodeType == XMLReader::ELEMENT && $reader->name == 'display-name') {
                            $channel_name = (string)$reader->readInnerXML();
                            $channels[$channel_id] = $channel_name;
                            break;
                        }
                    }
                    break;

                case 'programme':
                    $start = substr((string)$reader->getAttribute('start'), 0, -8);
                    $end = substr((string)$reader->getAttribute('stop'), 0, -8);
                    $channel = $channels[(int)$reader->getAttribute('channel')];
                    $title = '';
                    while ($reader->read()) {
                        if ($reader->nodeType == XMLReader::ELEMENT && $reader->name == 'title') {
                            $title = mb_strtolower((string)$reader->readInnerXML());
                            break;
                        }
                    }
                    if (strpos($title, $fraze) !== false) {
                        $founds[$channel][] = date("G:i", strtotime($start)) . ' - ' . date("G:i d.m.Y", strtotime($end)) . ' ' . (string)$reader->value;
                    }
                    break;
            }
        }
    }

    $reader->close();
    print_r($founds);
} else {
    echo "Usage: php index.php <search_word>\n";
}
