package libmozhi

func langListGemini(listType string) []List {
	langs := []List{{Id: "auto", Name: "Detect Language"}, {Id: "bg", Name: "Bulgarian"}, {Id: "cs", Name: "Czech"}, {Id: "da", Name: "Danish"}, {Id: "de", Name: "German"}, {Id: "el", Name: "Greek"}, {Id: "en", Name: "English"}, {Id: "es", Name: "Spanish"}, {Id: "et", Name: "Estonian"}, {Id: "fi", Name: "Finnish"}, {Id: "fr", Name: "French"}, {Id: "hu", Name: "Hungarian"}, {Id: "id", Name: "Indonesian"}, {Id: "it", Name: "Italian"}, {Id: "ja", Name: "Japanese"}, {Id: "ko", Name: "Korean"}, {Id: "lv", Name: "Latvian"}, {Id: "lt", Name: "Lithuanian"}, {Id: "nl", Name: "Dutch"}, {Id: "pl", Name: "Polish"}, {Id: "pt", Name: "Portuguese"}, {Id: "ro", Name: "Romanian"}, {Id: "ru", Name: "Russian"}, {Id: "sk", Name: "Slovak"}, {Id: "sl", Name: "Slovenian"}, {Id: "sv", Name: "Swedish"}, {Id: "tr", Name: "Turkish"}, {Id: "uk", Name: "Ukrainian"}, {Id: "zh", Name: "Chinese"}, {Id: "ar", Name: "Arabic"}}
	_ = listType
	return langs
}