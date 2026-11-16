package libmozhi

// Papago's languages (Naver Cloud docs, nov 2026). The web endpoint also
// translates pairs the official API lacks (e.g. ja-es, measured).
func langListPapago(listType string) []List {
	langs := []List{{Id: "auto", Name: "Detect Language"}, {Id: "zh-CN", Name: "Chinese (Simplified)"}, {Id: "zh-TW", Name: "Chinese (Traditional)"}, {Id: "en", Name: "English"}, {Id: "fr", Name: "French"}, {Id: "de", Name: "German"}, {Id: "id", Name: "Indonesian"}, {Id: "it", Name: "Italian"}, {Id: "ja", Name: "Japanese"}, {Id: "ko", Name: "Korean"}, {Id: "ru", Name: "Russian"}, {Id: "es", Name: "Spanish"}, {Id: "th", Name: "Thai"}, {Id: "vi", Name: "Vietnamese"}}
	_ = listType
	return langs
}
