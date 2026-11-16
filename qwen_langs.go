package libmozhi

// The same languages as the Cohere engine. Qwen is multilingual; English,
// Japanese and Spanish were tested (nov 2026).
func langListQwen(listType string) []List {
	langs := []List{{Id: "auto", Name: "Detect Language"}, {Id: "sq", Name: "Albanian"}, {Id: "ar", Name: "Arabic"}, {Id: "bn", Name: "Bengali"}, {Id: "bg", Name: "Bulgarian"}, {Id: "ca", Name: "Catalan"}, {Id: "zh", Name: "Chinese (Simplified)"}, {Id: "zh-TW", Name: "Chinese (Traditional)"}, {Id: "hr", Name: "Croatian"}, {Id: "cs", Name: "Czech"}, {Id: "da", Name: "Danish"}, {Id: "nl", Name: "Dutch"}, {Id: "en", Name: "English"}, {Id: "et", Name: "Estonian"}, {Id: "tl", Name: "Filipino"}, {Id: "fi", Name: "Finnish"}, {Id: "fr", Name: "French"}, {Id: "de", Name: "German"}, {Id: "el", Name: "Greek"}, {Id: "he", Name: "Hebrew"}, {Id: "hi", Name: "Hindi"}, {Id: "hu", Name: "Hungarian"}, {Id: "is", Name: "Icelandic"}, {Id: "id", Name: "Indonesian"}, {Id: "ga", Name: "Irish"}, {Id: "it", Name: "Italian"}, {Id: "ja", Name: "Japanese"}, {Id: "ko", Name: "Korean"}, {Id: "lv", Name: "Latvian"}, {Id: "lt", Name: "Lithuanian"}, {Id: "ms", Name: "Malay"}, {Id: "mt", Name: "Maltese"}, {Id: "no", Name: "Norwegian"}, {Id: "fa", Name: "Persian"}, {Id: "pl", Name: "Polish"}, {Id: "pt", Name: "Portuguese"}, {Id: "pa", Name: "Punjabi"}, {Id: "ro", Name: "Romanian"}, {Id: "ru", Name: "Russian"}, {Id: "sr", Name: "Serbian"}, {Id: "sk", Name: "Slovak"}, {Id: "sl", Name: "Slovenian"}, {Id: "es", Name: "Spanish"}, {Id: "sv", Name: "Swedish"}, {Id: "ta", Name: "Tamil"}, {Id: "te", Name: "Telugu"}, {Id: "th", Name: "Thai"}, {Id: "tr", Name: "Turkish"}, {Id: "uk", Name: "Ukrainian"}, {Id: "ur", Name: "Urdu"}, {Id: "vi", Name: "Vietnamese"}}
	_ = listType
	return langs
}
