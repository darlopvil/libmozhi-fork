package libmozhi

func langListReverso(listType string) []List {
	// IDs got from original simplytranslate-web and trial and error. Usually first three letters of language.
	ListData := []List{
		{
			Id:   "ar",
			Name: "Arabic",
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
			Id:   "fr",
			Name: "French",
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
			Id:   "per",
			Name: "Persian",
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
			Id:   "es",
			Name: "Spanish",
		},
		{
			Id:   "sv",
			Name: "Swedish",
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
			Id:   "uk",
			Name: "Ukrainian",
		},
	}
	return ListData
}
