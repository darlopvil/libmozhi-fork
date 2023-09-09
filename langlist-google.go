package libmozhi

func langListGoogle(listType string) []List {
	// IDs got from original simplytranslate-web and trial and error. Usually first three letters of language.
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
			Id:   "ay",
			Name: "Aymara",
		},
		{
			Id:   "az",
			Name: "Azerbaijani",
		},
		{
			Id:   "bm",
			Name: "Bambara",
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
			Id:   "bho",
			Name: "Bhojpuri",
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
			Id:   "ca",
			Name: "Catalan",
		},
		{
			Id:   "ceb",
			Name: "Cebuano",
		},
		{
			Id:   "ny",
			Name: "Chichewa",
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
			Id:   "co",
			Name: "Corsican",
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
			Id:   "dv",
			Name: "Dhivehi",
		},
		{
			Id:   "doi",
			Name: "Dogri",
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
			Id:   "eo",
			Name: "Esperanto",
		},
		{
			Id:   "et",
			Name: "Estonian",
		},
		{
			Id:   "ee",
			Name: "Ewe",
		},
		{
			Id:   "tl",
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
			Id:   "fy",
			Name: "Frisian",
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
			Id:   "gn",
			Name: "Guarani",
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
			Id:   "haw",
			Name: "Hawaiian",
		},
		{
			Id:   "iw",
			Name: "Hebrew",
		},
		{
			Id:   "hi",
			Name: "Hindi",
		},
		{
			Id:   "hmn",
			Name: "Hmong",
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
			Id:   "ilo",
			Name: "Ilocano",
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
			Id:   "jw",
			Name: "Javanese",
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
			Id:   "gom",
			Name: "Konkani",
		},
		{
			Id:   "ko",
			Name: "Korean",
		},
		{
			Id:   "kri",
			Name: "Krio",
		},
		{
			Id:   "ku",
			Name: "Kurdish (Kurmanji)",
		},
		{
			Id:   "ckb",
			Name: "Kurdish (Sorani)",
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
			Id:   "ln",
			Name: "Lingala",
		},
		{
			Id:   "lt",
			Name: "Lithuanian",
		},
		{
			Id:   "lg",
			Name: "Luganda",
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
			Id:   "mi",
			Name: "Maori",
		},
		{
			Id:   "mr",
			Name: "Marathi",
		},
		{
			Id:   "mni-Mtei",
			Name: "Meiteilon (Manipuri)",
		},
		{
			Id:   "lus",
			Name: "Mizo",
		},
		{
			Id:   "mn",
			Name: "Mongolian",
		},
		{
			Id:   "my",
			Name: "Myanmar (Burmese)",
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
			Id:   "or",
			Name: "Odia (Oriya)",
		},
		{
			Id:   "om",
			Name: "Oromo",
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
			Name: "Portuguese",
		},
		{
			Id:   "pa",
			Name: "Punjabi",
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
			Id:   "sa",
			Name: "Sanskrit",
		},
		{
			Id:   "gd",
			Name: "Scots Gaelic",
		},
		{
			Id:   "nso",
			Name: "Sepedi",
		},
		{
			Id:   "sr",
			Name: "Serbian",
		},
		{
			Id:   "st",
			Name: "Sesotho",
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
			Id:   "ti",
			Name: "Tigrinya",
		},
		{
			Id:   "ts",
			Name: "Tsonga",
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
			Id:   "ak",
			Name: "Twi",
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
			Id:   "ug",
			Name: "Uyghur",
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
			Id:   "yi",
			Name: "Yiddish",
		},
		{
			Id:   "yo",
			Name: "Yoruba",
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
