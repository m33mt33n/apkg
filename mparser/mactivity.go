package mparser

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
)

func handle_panic(pkg string) {
	a := recover()
	if a != nil {
		fmt.Printf("> error: %s: %s\n", pkg, a)
	}
}

func Get_main_activity(apk, pkg string) (string, uint8) {
	defer handle_panic(pkg)
	r, err := zip.OpenReader(apk)
	if err != nil {
		return "", 20
	}
	defer r.Close()
	for _, f := range r.File {
		if f.Name == "AndroidManifest.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", 21
			}
			xml_str, err := DecompressXML(rc)
			if err != nil {
				return "", 22
			}
			xml_bytes := []byte(xml_str)
			var mf Manifest
			err = xml.Unmarshal(xml_bytes, &mf)
			activites := mf.GetActivites()
			activity_found := false
			activity_name := ""
			for _, activity := range activites {
				if activity_found {
					break
				}
				for _, intent_filter := range activity.IntentFilters {
					if activity_found {
						break
					}
					for _, action := range intent_filter.Actions {
						if activity_found {
							break
						}
						if action.Name == "android.intent.action.MAIN" {
							for _, category := range intent_filter.Categories {
								if category.Name == "android.intent.category.LAUNCHER" {
									activity_name = activity.Name
									activity_found = true
									break
								}
							}
						}
					}
				}
				if !activity_found {
					activity_aliases := mf.GetActivitesAliases()
					for _, alias := range activity_aliases {
						if activity_found {
							break
						}
						for _, intent_filter := range alias.IntentFilters {
							if activity_found {
								break
							}
							for _, action := range intent_filter.Actions {
								if activity_found {
									break
								}
								if action.Name == "android.intent.action.MAIN" {
									for _, category := range intent_filter.Categories {
										if category.Name == "android.intent.category.LAUNCHER" {
											activity_name = alias.Name
											activity_found = true
											break
										}
									}
								}
							}
						}
					}
				}
			}
			rc.Close()
			return activity_name, 0
			break
		}
	}
	return "", 23
}
