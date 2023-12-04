## DeepL
curl 'https://dict.deepl.com/english-french/search?ajax=1&source=english&onlyDictEntries=1&translator=dnsof7h3k2lgh3gda&kind=full&eventkind=keyup&forleftside=true&il=en' -XPOST -H 'Content-Type: application/x-www-form-urlencoded' -d 'query=SOMETHING' 
alternativewords in struct given by gdeeplx

## Google
Pretty complex, see: https://codeberg.org/SimpleWeb/SimplyTranslate-Engines/src/branch/master/simplytranslate_engines/googletranslate.py

## Reverso
Set contextResults option true in reverso json:
curl -XPOST -H "Content-type: application/json" -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; rv:110.0) Gecko/20100101 Firefox/110.0' https://api.reverso.net/translate/v1/translation -d '{ "format": "text", "from": "en", "to": "fr", "input":"hello", "options": {"sentenceSplitter": false, "origin":"translation.web", contextResults: true, languageDetection: true} }'
AND
curl https://synonyms.reverso.net/api/v2//search/en/test?limit=60&merge=true&rude=false&colloquial=false&exact=true

## Yandex
Detect Lang as well

## Watson
No support

## LibreTranslate
No support

## MyMemory
No support

## DuckDuckGo
No support
