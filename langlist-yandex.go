package libmozhi

func langListYandex(listType string) []List {
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
			Id:   "az",
			Name: "Azerbaijani",
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
			Id:   "be",
			Name: "Belarusian",
		},
		{
			Id:   "bn",
			Name: "Bengali",
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
			Id:   "my",
			Name: "Burmese",
		},
		{
			Id:   "ca",
			Name: "Catalan",
		},
		{
			Id:   "ceb",
			Name: "Cebuano",
		},
		{
			Id:   "zh",
			Name: "Chinese (Simplified)",
		},
		{
			Id:   "cv",
			Name: "Chuvash",
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
			Id:   "nl",
			Name: "Dutch",
		},
		{
			Id:   "sjn",
			Name: "Elvish (Sindarin)",
		},
		{
			Id:   "emj",
			Name: "Emoji",
		},
		{
			Id:   "en",
			Name: "English",
		},
		{
			Id:   "eo",
			Name: "Esperanto",
		},
		{
			Id:   "et",
			Name: "Estonian",
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
			Id:   "gl",
			Name: "Galician",
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
			Name: "Haitian",
		},
		{
			Id:   "he",
			Name: "Hebrew",
		},
		{
			Id:   "mrj",
			Name: "Hill Mari",
		},
		{
			Id:   "hi",
			Name: "Hindi",
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
			Id:   "id",
			Name: "Indonesian",
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
			Id:   "jv",
			Name: "Javanese",
		},
		{
			Id:   "kn",
			Name: "Kannada",
		},
		{
			Id:   "kazlat",
			Name: "Kazakh (Latin)",
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
			Id:   "ko",
			Name: "Korean",
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
			Id:   "la",
			Name: "Latin",
		},
		{
			Id:   "lv",
			Name: "Latvian",
		},
		{
			Id:   "lt",
			Name: "Lithuanian",
		},
		{
			Id:   "lb",
			Name: "Luxembourgish",
		},
		{
			Id:   "mk",
			Name: "Macedonian",
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
			Id:   "mi",
			Name: "Maori",
		},
		{
			Id:   "mr",
			Name: "Marathi",
		},
		{
			Id:   "mhr",
			Name: "Mari",
		},
		{
			Id:   "mn",
			Name: "Mongolian",
		},
		{
			Id:   "ne",
			Name: "Nepali",
		},
		{
			Id:   "no",
			Name: "Norwegian",
		},
		{
			Id:   "pap",
			Name: "Papiamento",
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
			Id:   "pt-BR",
			Name: "Portuguese (Brazilian)",
		},
		{
			Id:   "pt",
			Name: "Portuguese",
		},
		{
			Id:   "pa",
			Name: "Punjabi",
		},
		{
			Id:   "ro",
			Name: "Romanian",
		},
		{
			Id:   "ru",
			Name: "Russian",
		},
		{
			Id:   "gd",
			Name: "Scottish Gaelic",
		},
		{
			Id:   "sr-Latn",
			Name: "Serbian (Latin)",
		},
		{
			Id:   "sr",
			Name: "Serbian",
		},
		{
			Id:   "si",
			Name: "Sinhalese",
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
			Id:   "es",
			Name: "Spanish",
		},
		{
			Id:   "su",
			Name: "Sundanese",
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
			Id:   "tl",
			Name: "Tagalog",
		},
		{
			Id:   "tg",
			Name: "Tajik",
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
			Id:   "tr",
			Name: "Turkish",
		},
		{
			Id:   "udm",
			Name: "Udmurt",
		},
		{
			Id:   "uk",
			Name: "Ukrainian",
		},
		{
			Id:   "ur",
			Name: "Urdu",
		},
		{
			Id:   "uzbcyr",
			Name: "Uzbek (Cyrillic)",
		},
		{
			Id:   "uz",
			Name: "Uzbek",
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
			Id:   "sah",
			Name: "Yakut",
		},
		{
			Id:   "yi",
			Name: "Yiddish",
		},
		{
			Id:   "zu",
			Name: "Zulu",
		},
	}
	return ListData
}
