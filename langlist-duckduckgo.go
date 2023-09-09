package libmozhi

func langListDuckDuckGo(listType string) []List {
	// DuckDuckGo is just Bing translate but easier to scrape :)
	// Converted json of https://github.com/plainheart/bing-translate-api/blob/master/src/lang.json to this format
	ListData := []List{
		{
			Id:   "af",
			Name: "Afrikaans",
		},
		{
			Id:   "sq",
			Name: "Albanian",
		},
		{
			Id:   "am",
			Name: "Amharic",
		},
		{
			Id:   "ar",
			Name: "Arabic",
		},
		{
			Id:   "hy",
			Name: "Armenian",
		},
		{
			Id:   "as",
			Name: "Assamese",
		},
		{
			Id:   "az",
			Name: "Azerbaijani",
		},
		{
			Id:   "bn",
			Name: "Bangla",
		},
		{
			Id:   "ba",
			Name: "Bashkir",
		},
		{
			Id:   "eu",
			Name: "Basque",
		},
		{
			Id:   "bs",
			Name: "Bosnian",
		},
		{
			Id:   "bg",
			Name: "Bulgarian",
		},
		{
			Id:   "yue",
			Name: "Cantonese (Traditional)",
		},
		{
			Id:   "ca",
			Name: "Catalan",
		},
		{
			Id:   "lzh",
			Name: "Chinese (Literary)",
		},
		{
			Id:   "zh-Hans",
			Name: "Chinese Simplified",
		},
		{
			Id:   "zh-Hant",
			Name: "Chinese Traditional",
		},
		{
			Id:   "hr",
			Name: "Croatian",
		},
		{
			Id:   "cs",
			Name: "Czech",
		},
		{
			Id:   "da",
			Name: "Danish",
		},
		{
			Id:   "prs",
			Name: "Dari",
		},
		{
			Id:   "dv",
			Name: "Divehi",
		},
		{
			Id:   "nl",
			Name: "Dutch",
		},
		{
			Id:   "en",
			Name: "English",
		},
		{
			Id:   "et",
			Name: "Estonian",
		},
		{
			Id:   "fo",
			Name: "Faroese",
		},
		{
			Id:   "fj",
			Name: "Fijian",
		},
		{
			Id:   "fil",
			Name: "Filipino",
		},
		{
			Id:   "fi",
			Name: "Finnish",
		},
		{
			Id:   "fr",
			Name: "French",
		},
		{
			Id:   "fr-CA",
			Name: "French (Canada)",
		},
		{
			Id:   "gl",
			Name: "Galician",
		},
		{
			Id:   "lug",
			Name: "Ganda",
		},
		{
			Id:   "ka",
			Name: "Georgian",
		},
		{
			Id:   "de",
			Name: "German",
		},
		{
			Id:   "el",
			Name: "Greek",
		},
		{
			Id:   "gu",
			Name: "Gujarati",
		},
		{
			Id:   "ht",
			Name: "Haitian Creole",
		},
		{
			Id:   "ha",
			Name: "Hausa",
		},
		{
			Id:   "he",
			Name: "Hebrew",
		},
		{
			Id:   "hi",
			Name: "Hindi",
		},
		{
			Id:   "mww",
			Name: "Hmong Daw",
		},
		{
			Id:   "hu",
			Name: "Hungarian",
		},
		{
			Id:   "is",
			Name: "Icelandic",
		},
		{
			Id:   "ig",
			Name: "Igbo",
		},
		{
			Id:   "id",
			Name: "Indonesian",
		},
		{
			Id:   "ikt",
			Name: "Inuinnaqtun",
		},
		{
			Id:   "iu",
			Name: "Inuktitut",
		},
		{
			Id:   "iu-Latn",
			Name: "Inuktitut (Latin)",
		},
		{
			Id:   "ga",
			Name: "Irish",
		},
		{
			Id:   "it",
			Name: "Italian",
		},
		{
			Id:   "ja",
			Name: "Japanese",
		},
		{
			Id:   "kn",
			Name: "Kannada",
		},
		{
			Id:   "kk",
			Name: "Kazakh",
		},
		{
			Id:   "km",
			Name: "Khmer",
		},
		{
			Id:   "rw",
			Name: "Kinyarwanda",
		},
		{
			Id:   "tlh-Latn",
			Name: "Klingon (Latin)",
		},
		{
			Id:   "gom",
			Name: "Konkani",
		},
		{
			Id:   "ko",
			Name: "Korean",
		},
		{
			Id:   "ku",
			Name: "Kurdish (Central)",
		},
		{
			Id:   "kmr",
			Name: "Kurdish (Northern)",
		},
		{
			Id:   "ky",
			Name: "Kyrgyz",
		},
		{
			Id:   "lo",
			Name: "Lao",
		},
		{
			Id:   "lv",
			Name: "Latvian",
		},
		{
			Id:   "ln",
			Name: "Lingala",
		},
		{
			Id:   "lt",
			Name: "Lithuanian",
		},
		{
			Id:   "dsb",
			Name: "Lower Sorbian",
		},
		{
			Id:   "mk",
			Name: "Macedonian",
		},
		{
			Id:   "mai",
			Name: "Maithili",
		},
		{
			Id:   "mg",
			Name: "Malagasy",
		},
		{
			Id:   "ms",
			Name: "Malay",
		},
		{
			Id:   "ml",
			Name: "Malayalam",
		},
		{
			Id:   "mt",
			Name: "Maltese",
		},
		{
			Id:   "mr",
			Name: "Marathi",
		},
		{
			Id:   "mn-Cyrl",
			Name: "Mongolian (Cyrillic)",
		},
		{
			Id:   "mn-Mong",
			Name: "Mongolian (Traditional)",
		},
		{
			Id:   "my",
			Name: "Myanmar (Burmese)",
		},
		{
			Id:   "mi",
			Name: "Māori",
		},
		{
			Id:   "ne",
			Name: "Nepali",
		},
		{
			Id:   "nb",
			Name: "Norwegian",
		},
		{
			Id:   "nya",
			Name: "Nyanja",
		},
		{
			Id:   "or",
			Name: "Odia",
		},
		{
			Id:   "ps",
			Name: "Pashto",
		},
		{
			Id:   "fa",
			Name: "Persian",
		},
		{
			Id:   "pl",
			Name: "Polish",
		},
		{
			Id:   "pt",
			Name: "Portuguese (Brazil)",
		},
		{
			Id:   "pt-PT",
			Name: "Portuguese (Portugal)",
		},
		{
			Id:   "pa",
			Name: "Punjabi",
		},
		{
			Id:   "otq",
			Name: "Querétaro Otomi",
		},
		{
			Id:   "ro",
			Name: "Romanian",
		},
		{
			Id:   "run",
			Name: "Rundi",
		},
		{
			Id:   "ru",
			Name: "Russian",
		},
		{
			Id:   "sm",
			Name: "Samoan",
		},
		{
			Id:   "sr-Cyrl",
			Name: "Serbian (Cyrillic)",
		},
		{
			Id:   "sr-Latn",
			Name: "Serbian (Latin)",
		},
		{
			Id:   "st",
			Name: "Sesotho",
		},
		{
			Id:   "nso",
			Name: "Sesotho sa Leboa",
		},
		{
			Id:   "tn",
			Name: "Setswana",
		},
		{
			Id:   "sn",
			Name: "Shona",
		},
		{
			Id:   "sd",
			Name: "Sindhi",
		},
		{
			Id:   "si",
			Name: "Sinhala",
		},
		{
			Id:   "sk",
			Name: "Slovak",
		},
		{
			Id:   "sl",
			Name: "Slovenian",
		},
		{
			Id:   "so",
			Name: "Somali",
		},
		{
			Id:   "es",
			Name: "Spanish",
		},
		{
			Id:   "sw",
			Name: "Swahili",
		},
		{
			Id:   "sv",
			Name: "Swedish",
		},
		{
			Id:   "ty",
			Name: "Tahitian",
		},
		{
			Id:   "ta",
			Name: "Tamil",
		},
		{
			Id:   "tt",
			Name: "Tatar",
		},
		{
			Id:   "te",
			Name: "Telugu",
		},
		{
			Id:   "th",
			Name: "Thai",
		},
		{
			Id:   "bo",
			Name: "Tibetan",
		},
		{
			Id:   "ti",
			Name: "Tigrinya",
		},
		{
			Id:   "to",
			Name: "Tongan",
		},
		{
			Id:   "tr",
			Name: "Turkish",
		},
		{
			Id:   "tk",
			Name: "Turkmen",
		},
		{
			Id:   "uk",
			Name: "Ukrainian",
		},
		{
			Id:   "hsb",
			Name: "Upper Sorbian",
		},
		{
			Id:   "ur",
			Name: "Urdu",
		},
		{
			Id:   "ug",
			Name: "Uyghur",
		},
		{
			Id:   "uz",
			Name: "Uzbek (Latin)",
		},
		{
			Id:   "vi",
			Name: "Vietnamese",
		},
		{
			Id:   "cy",
			Name: "Welsh",
		},
		{
			Id:   "xh",
			Name: "Xhosa",
		},
		{
			Id:   "yo",
			Name: "Yoruba",
		},
		{
			Id:   "yua",
			Name: "Yucatec Maya",
		},
		{
			Id:   "zu",
			Name: "Zulu",
		},
	}
	var ListD []List
	if listType == "sl" {
		auto := []List{{
			Id:   "auto",
			Name: "Detect Language",
		}}
		ListD = append(append([]List{}, auto...), ListData...)
	} else {
		ListD = ListData
	}
	return ListD
}
