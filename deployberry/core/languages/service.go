package languages

import (
	"shared/repository"
)

func CheckAllService() (map[string]any, error) {
	db := repository.GetDB()
	var languages []repository.Language

	err := db.Find(&languages).Error
	if err != nil {
		return nil, err
	}

	// Format the response data - return clean semantic versions only
	data := map[string]any{
		"php":    "",
		"node":   "",
		"python": "",
		"golang": "",
	}

	// Update with installed and active versions (already normalized in DB)
	for _, lang := range languages {
		if lang.Active && lang.Version != "" {
			// Versions in database should already be normalized (x.x.x format)
			// Frontend should receive clean versions directly
			data[lang.Name] = lang.Version
		}
	}

	return data, nil
}

type LanguageInfo struct {
	Version   string `json:"version"`
	Installed bool   `json:"installed"`
	Active    bool   `json:"active"`
}

func CheckOneService(lang string) (LanguageInfo, error) {
	db := repository.GetDB()
	var language repository.Language
	err := db.Where("name = ?", lang).First(&language).Error
	if err != nil {
		return LanguageInfo{}, err
	}

	return LanguageInfo{
		Version:   language.Version,
		Installed: language.Version != "",
		Active:    language.Active,
	}, nil
}
