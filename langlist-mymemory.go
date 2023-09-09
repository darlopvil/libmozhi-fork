package libmozhi

func langListMyMemory(listType string) []List {
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
			Id:   "bjs",
			Name: "Bajan",
		},
		{
			Id:   "rm",
			Name: "Balkan Gipsy",
		},
		{
			Id:   "eu",
			Name: "Basque",
		},
		{
			Id:   "bem",
			Name: "Bemba",
		},
		{
			Id:   "bn",
			Name: "Bengali",
		},
		{
			Id:   "be",
			Name: "Bielarus",
		},
		{
			Id:   "bi",
			Name: "Bislama",
		},
		{
			Id:   "bs",
			Name: "Bosnian",
		},
		{
			Id:   "br",
			Name: "Breton",
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
			Id:   "ch",
			Name: "Chamorro",
		},
		{
			Id:   "zh",
			Name: "Chinese (Simplified)",
		},
		{
			Id:   "zh-TW",
			Name: "Chinese (Traditional)",
		},
		{
			Id:   "zdj",
			Name: "Comorian",
		},
		{
			Id:   "cop",
			Name: "Coptic",
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
			Id:   "dz",
			Name: "Dzongkha",
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
			Id:   "fnG",
			Name: "Fanagalo",
		},
		{
			Id:   "fo",
			Name: "Faroese",
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
			Id:   "grc",
			Name: "Greek (Classical)",
		},
		{
			Id:   "gu",
			Name: "Gujarati",
		},
		{
			Id:   "ha",
			Name: "Hausa",
		},
		{
			Id:   "haw",
			Name: "Hawaiian",
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
			Id:   "kl",
			Name: "Inuktitut (Greenland)",
		},
		{
			Id:   "ga",
			Name: "Irish Gaelic",
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
			Id:   "kea",
			Name: "Kabuverdianu",
		},
		{
			Id:   "kab",
			Name: "Kabylian",
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
			Id:   "rn",
			Name: "Kirundi",
		},
		{
			Id:   "ko",
			Name: "Korean",
		},
		{
			Id:   "ku",
			Name: "Kurdish",
		},
		{
			Id:   "ckb",
			Name: "Kurdish Sorani",
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
			Id:   "dv",
			Name: "Maldivian",
		},
		{
			Id:   "mt",
			Name: "Maltese",
		},
		{
			Id:   "gv",
			Name: "Manx Gaelic",
		},
		{
			Id:   "mi",
			Name: "Maori",
		},
		{
			Id:   "mh",
			Name: "Marshallese",
		},
		{
			Id:   "men",
			Name: "Mende",
		},
		{
			Id:   "mn",
			Name: "Mongolian",
		},
		{
			Id:   "mfe",
			Name: "Morisyen",
		},
		{
			Id:   "ne",
			Name: "Nepali",
		},
		{
			Id:   "niu",
			Name: "Niuean",
		},
		{
			Id:   "no",
			Name: "Norwegian",
		},
		{
			Id:   "ny",
			Name: "Nyanja",
		},
		{
			Id:   "ur",
			Name: "Pakistani",
		},
		{
			Id:   "pau",
			Name: "Palauan",
		},
		{
			Id:   "pa",
			Name: "Panjabi",
		},
		{
			Id:   "pap",
			Name: "Papiamentu",
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
			Id:   "pis",
			Name: "Pijin",
		},
		{
			Id:   "pl",
			Name: "Polish",
		},
		{
			Id:   "pt",
			Name: "Portuguese",
		},
		{
			Id:   "pot",
			Name: "Potawatomi",
		},
		{
			Id:   "qu",
			Name: "Quechua",
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
			Id:   "sm",
			Name: "Samoan",
		},
		{
			Id:   "sg",
			Name: "Sango",
		},
		{
			Id:   "gd",
			Name: "Scots Gaelic",
		},
		{
			Id:   "sr",
			Name: "Serbian",
		},
		{
			Id:   "sn",
			Name: "Shona",
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
			Id:   "st",
			Name: "Sotho Southern",
		},
		{
			Id:   "es",
			Name: "Spanish",
		},
		{
			Id:   "srn",
			Name: "Sranan Tongo",
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
			Id:   "syc",
			Name: "Syriac (Aramic)",
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
			Id:   "tmh",
			Name: "Tamashek (Tuareg)",
		},
		{
			Id:   "ta",
			Name: "Tamil",
		},
		{
			Id:   "te",
			Name: "Telugu",
		},
		{
			Id:   "tet",
			Name: "Tetum",
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
			Id:   "tpi",
			Name: "Tok Pisin",
		},
		{
			Id:   "tkl",
			Name: "Tokelauan",
		},
		{
			Id:   "to",
			Name: "Tongan",
		},
		{
			Id:   "tn",
			Name: "Tswana",
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
			Id:   "tvl",
			Name: "Tuvaluan",
		},
		{
			Id:   "uk",
			Name: "Ukrainian",
		},
		{
			Id:   "ppk",
			Name: "Uma",
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
			Id:   "wls",
			Name: "Wallisian",
		},
		{
			Id:   "cy",
			Name: "Welsh",
		},
		{
			Id:   "wo",
			Name: "Wolof",
		},
		{
			Id:   "xh",
			Name: "Xhosa",
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
