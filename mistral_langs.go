package libmozhi

func langListMistral(listType string) []List {
	langs := []List{{Id: "auto", Name: "Detect Language"}, {Id: "ar", Name: "Arabic"}, {Id: "bg", Name: "Bulgarian"}, {Id: "cs", Name: "Czech"}, {Id: "da", Name: "Danish"}, {Id: "de", Name: "German"}, {Id: "el", Name: "Greek"}, {Id: "en", Name: "English"}, {Id: "es", Name: "Spanish"}, {Id: "fi", Name: "Finnish"}, {Id: "fr", Name: "French"}, {Id: "hi", Name: "Hindi"}, {Id: "hu", Name: "Hungarian"}, {Id: "id", Name: "Indonesian"}, {Id: "it", Name: "Italian"}, {Id: "ja", Name: "Japanese"}, {Id: "ko", Name: "Korean"}, {Id: "nl", Name: "Dutch"}, {Id: "pl", Name: "Polish"}, {Id: "pt", Name: "Portuguese"}, {Id: "ro", Name: "Romanian"}, {Id: "ru", Name: "Russian"}, {Id: "sk", Name: "Slovak"}, {Id: "sv", Name: "Swedish"}, {Id: "tr", Name: "Turkish"}, {Id: "uk", Name: "Ukrainian"}, {Id: "vi", Name: "Vietnamese"}, {Id: "zh", Name: "Chinese"}}
	_ = listType
	return langs
}