package libmozhi

func langListDeepl(listType string) []List {
	// IDs got from deepl.com/translator
	// Every time you change language, the # will get updated with the lang code.
	ListData := []List{
		{
			Id:   "bg",
			Name: "Bulgarian",
		},
		{
			Id:   "zh",
			Name: "Chinese (Simplified)",
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
			Id:   "en",
			Name: "English",
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
			Id:   "de",
			Name: "Germany",
		},
		{
			Id:   "el",
			Name: "Greek",
		},
		{
			Id:   "hu",
			Name: "Hungarian",
		},
		{
			Id:   "id",
			Name: "Indonesian",
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
			Id:   "ko",
			Name: "Korean",
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
			Id:   "nb",
			Name: "Norwegian",
		},
		{
			Id:   "pl",
			Name: "Polish",
		},
		{
			Id:   "pt",
			Name: "Portugese",
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
			Id:   "sv",
			Name: "Swedish",
		},
		{
			Id:   "tr",
			Name: "Turkish",
		},
		{
			Id:   "uk",
			Name: "Ukrainian",
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
