package libmozhi

func langListWatson(listType string) []List {
	// IDs got from https://www.loc.gov/standards/iso639-2/php/code_list.php and tested to make sure they work. Exceptions fr-CA zh-CN/TW
	ListData := []List{
		{
			Id:   "ar",
			Name: "Arabic",
		},
		{
			Id:   "ba",
			Name: "Basque",
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
			Id:   "ca",
			Name: "Catalan",
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
			Id:   "fr-CA",
			Name: "Canadian French",
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
			Id:   "kn",
			Name: "Kannada",
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
			Id:   "cnr",
			Name: "Montenegrin",
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
			Id:   "pl",
			Name: "Polish",
		},
		{
			Id:   "pt",
			Name: "Portugese",
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
			Id:   "sv",
			Name: "Swedish",
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
		{
			Id:   "ur",
			Name: "Urdu",
		},
		{
			Id:   "vi",
			Name: "Vietnamese",
		},
		{
			Id:   "cy",
			Name: "Welsh",
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
