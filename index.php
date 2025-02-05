<?php

if (isset($argv[1])) {
    $file = "epg2.xml.gz";
    //check file date
    if (!file_exists($file) || filemtime("epg2.xml.gz") < time() - 86400) {
        //load file from web
        $url = "http://epg.one/epg2.xml.gz";
        file_put_contents($file, file_get_contents($url));
    }
    
    $xml = simplexml_load_file("compress.zlib://{$file}");
    $channels = [];
    foreach ($xml->channel as $c) {
        $channels[ $c['id']->__toString() ] = $c->{'display-name'}->__toString();
    }
    $fraze = mb_strtolower($argv[1]);
    $founds = [];
    foreach ($xml->programme as $item) {
        $title = mb_strtolower($item->title);
        if (strpos($title, $fraze) !== false) {
            $start = substr((string)$item["start"], 0, -8);
            $end   = substr((string)$item["stop"], 0, -8);
            $founds[$channels[(int)$item['channel']]][] = date("G:i", strtotime($start)).' - '.date("G:i d.m.Y", strtotime($end)).' '.$item->title->__toString();
        }
    }
    
    print_r($founds);
} else {
    echo "Usage: php index.php <search_word>\n";
}
