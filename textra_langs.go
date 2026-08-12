package libmozhi

func langListTexTra(listType string) []List {
	langs := []List{{Id: "ja", Name: "Japanese"}, {Id: "en", Name: "English"}, {Id: "zh", Name: "Chinese"}, {Id: "ko", Name: "Korean"}, {Id: "fr", Name: "French"}, {Id: "es", Name: "Spanish"}, {Id: "de", Name: "German"}, {Id: "it", Name: "Italian"}, {Id: "pt", Name: "Portuguese"}, {Id: "ru", Name: "Russian"}, {Id: "th", Name: "Thai"}, {Id: "vi", Name: "Vietnamese"}, {Id: "id", Name: "Indonesian"}, {Id: "my", Name: "Burmese"}}
	_ = listType
	return langs
}