// an orthography conversion tool for mi'kmaw
// most lines here are using the strings package to replace characters
// works by converting inputs to a "unified orthography" and then converting back out into different orthographies
// unified orthography uses several characters to make things easier:
// "*" is schwa
// "@" is /a:/
// "3" is /e:/
// "!" is /i:/
// "%" is /o:/
// "&" is /u:/
// "$" is for /kw/ (coarticulated)
// "#" is for /gw/ (coarticulated, intervocalically voiced)
// "8" is syllabic /m/, non-initial
// "9" is syllabic /n/, non-initial
// "0" is syllabic /l/, non-initial
// "6" is initial syllabic /l/
// "7" is initial syllabic /n/
// "j" is voiced /dʒ/
// "c" is voiceless /tʃ/
// many unique sequences are used for rand orthography special characters, in case the user cannot type them. these can be found in the character substitution table

package converter

import (
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

type Output struct { // the forms to be output
	FrancisSmith        string
	Listuguj            string
	Pacifique           string
	PacifiqueDisclaimer bool
	Rand                string
	RandDisclaimer      bool
	Lexicon             string
	Metallic            string
}

func ConverterInit() error {
	http.HandleFunc("/convert", orthoIndexHandler) // create the webpage
	return nil
}

func orthoIndexHandler(writer http.ResponseWriter, reader *http.Request) {
	var normalStr string // the "normal" string is that which is normalized for conversion to any orthography
	var wasUpper bool
	var PacifiqueDisclaimer bool = false
	var RandDisclaimer bool = false
	if reader.Method == http.MethodPost { // if the "go" button is pressed
		InputStr := reader.FormValue("wordinput")              // get the input string
		orthographyChoice := reader.FormValue("orthographies") // a string value correstponding to the orthography chosen by the user
		if InputStr != "" {                                    // if the input is not empty
			if strings.ToUpper(string([]rune(InputStr)[0])) == string([]rune(InputStr)[0]) {
				wasUpper = true
			}
			InputStr = strings.ToLower(InputStr)
			switch orthographyChoice {
			case "francissmith":
				normalStr = normalizeFrancisSmith(InputStr)
			case "listuguj":
				normalStr = normalizeListuguj(InputStr)
			case "pacifique":
				normalStr = normalizePacifique(InputStr)
				PacifiqueDisclaimer = true
			case "rand":
				normalStr = normalizeRand(InputStr)
				RandDisclaimer = true
			case "lexicon":
				normalStr = normalizeLexicon(InputStr)
			case "metallic":
				normalStr = normalizeMetallic(InputStr)
			default:
				fmt.Println("orthography type missing")
			}
		}
	} else { // if the button was not pressed (i.e. on first load of the page without cache)
		normalStr = normalizeFrancisSmith("put*p") // default is "put*p"
	}

	OutputWords := encodeOutput(normalStr)
	OutputWords.PacifiqueDisclaimer = PacifiqueDisclaimer
	OutputWords.RandDisclaimer = RandDisclaimer
	if wasUpper && len(normalStr) > 1 {
		francisSmithRunes := []rune(OutputWords.FrancisSmith)
		OutputWords.FrancisSmith = strings.ToUpper(string(francisSmithRunes[0])) + string(francisSmithRunes[1:])
		listugujRunes := []rune(OutputWords.Listuguj)
		OutputWords.Listuguj = strings.ToUpper(string(listugujRunes[0])) + string(listugujRunes[1:])
		pacifiqueRunes := []rune(OutputWords.Pacifique)
		OutputWords.Pacifique = strings.ToUpper(string(pacifiqueRunes[0])) + string(pacifiqueRunes[1:])
		randRunes := []rune(OutputWords.Rand)
		OutputWords.Rand = strings.ToUpper(string(randRunes[0])) + string(randRunes[1:])
		lexiconRunes := []rune(OutputWords.Lexicon)
		OutputWords.Lexicon = strings.ToUpper(string(lexiconRunes[0])) + string(lexiconRunes[1:])
		metallicRunes := []rune(OutputWords.Metallic)
		OutputWords.Metallic = strings.ToUpper(string(metallicRunes[0])) + string(metallicRunes[1:])
	}

	template, templateBuildErr := template.ParseFiles("converter/convertertemplate.html.temp") // parse conjugatortemplate.html.temp
	if templateBuildErr != nil {                                                               // if an error is thrown
		fmt.Println(templateBuildErr)
	}
	template.Execute(writer, OutputWords) // execute the template
}

func IsConsonant(category string) bool { // returns true if the passed slice is in this list
	switch category {
	case
		"j",
		"c",
		"k",
		"g",
		"p",
		"b",
		"q",
		"s",
		"t",
		"d",
		"w",
		"y":
		return true
	}
	return false
}

func IsSonorant(category string) bool { // returns true if the passed slice is in this list
	switch category {
	case
		"l",
		"m",
		"n",
		"6",
		"7",
		"8",
		"9",
		"0":
		return true
	}
	return false
}

func IsSemivowel(category string) bool { // returns true if the passed slice is in this list
	switch category {
	case
		"w",
		"y":
		return true
	}
	return false
}

func IsAllophonicallyVoiced(category string) bool { // returns true if the passed slice is in this list
	switch category {
	case
		"j",
		"k",
		"$",
		"p",
		"t":
		return true
	}
	return false
}

func normalizeFrancisSmith(inputStr string) string {
	outputStr := inputStr
	outputStr = strings.Replace(outputStr, "ɨ", "*", -1)
	outputStr = strings.Replace(outputStr, "a'", "@", -1)
	outputStr = strings.Replace(outputStr, "e'", "3", -1)
	outputStr = strings.Replace(outputStr, "i'", "!", -1)
	outputStr = strings.Replace(outputStr, "o'", "%", -1)
	outputStr = strings.Replace(outputStr, "u'", "&", -1)
	outputStr = strings.Replace(outputStr, "kw", "$", -1)

	for charIndex, character := range outputStr {
		if IsSonorant(string(character)) {
			if charIndex != 0 {
				if (IsConsonant(string(outputStr[charIndex-1])) || IsSonorant(string(outputStr[charIndex-1]))) &&
					(!IsSemivowel(string(outputStr[charIndex-1]))) {
					if string(character) == "m" {
						outputStr = string(outputStr[:charIndex]) + "8" + outputStr[charIndex+1:]
					} else if string(character) == "n" {
						outputStr = string(outputStr[:charIndex]) + "9" + outputStr[charIndex+1:]
					} else if string(character) == "l" {
						outputStr = string(outputStr[:charIndex]) + "0" + outputStr[charIndex+1:]
					}
				}
			}
		} else if IsAllophonicallyVoiced(string(character)) {
			if charIndex == 0 {
				if !(IsConsonant(string(outputStr[charIndex+1]))) {
					if string(character) == "t" {
						outputStr = "d" + outputStr[charIndex+1:]
					} else if string(character) == "p" {
						outputStr = "b" + outputStr[charIndex+1:]
					} else if string(character) == "k" {
						outputStr = "g" + outputStr[charIndex+1:]
					} else if string(character) == "$" {
						outputStr = "#" + outputStr[charIndex+1:]
					}
				} else if IsConsonant(string(outputStr[charIndex+1])) && string(character) == "j" {
					outputStr = "c" + outputStr[charIndex+1:]
				}
			} else if charIndex != len(outputStr)-1 && string(outputStr[charIndex+1]) != " " {
				if !(IsConsonant(string(outputStr[charIndex-1]))) && !(IsConsonant(string(outputStr[charIndex+1]))) {
					if string(character) == "t" {
						outputStr = string(outputStr[:charIndex]) + "d" + outputStr[charIndex+1:]
					} else if string(character) == "p" {
						outputStr = string(outputStr[:charIndex]) + "b" + outputStr[charIndex+1:]
					} else if string(character) == "k" {
						outputStr = string(outputStr[:charIndex]) + "g" + outputStr[charIndex+1:]
					} else if string(character) == "$" {
						outputStr = string(outputStr[:charIndex]) + "#" + outputStr[charIndex+1:]
					}
				} else if (IsConsonant(string(outputStr[charIndex-1])) || IsConsonant(string(outputStr[charIndex+1]))) && string(character) == "j" {
					outputStr = string(outputStr[:charIndex]) + "c" + outputStr[charIndex+1:]
				}
			} else if charIndex == len(outputStr)-1 || string(outputStr[charIndex+1]) == " " {
				if string(character) == "j" {
					outputStr = outputStr[:charIndex] + "c"
				}
			}
		}
	}

	if IsConsonant(string((outputStr[0]))) && IsConsonant(string((outputStr[1]))) {
		outputStr = "*" + outputStr
	}

	outputStr = strings.Replace(outputStr, "l'", "6", -1)
	outputStr = strings.Replace(outputStr, "n'", "7", -1)

	// sonorants after syllabic word-initial sonorants do not need to be recognized as such
	outputStr = strings.Replace(outputStr, "68", "6m", -1)
	outputStr = strings.Replace(outputStr, "69", "6n", -1)
	outputStr = strings.Replace(outputStr, "60", "6l", -1)
	outputStr = strings.Replace(outputStr, "78", "7m", -1)
	outputStr = strings.Replace(outputStr, "79", "7n", -1)
	outputStr = strings.Replace(outputStr, "70", "7l", -1)

	return outputStr
}

func normalizeListuguj(inputStr string) string {
	outputStr := inputStr

	outputStr = strings.Replace(outputStr, "ai", "ay", -1)
	outputStr = strings.Replace(outputStr, "a'i", "@y", -1)
	outputStr = strings.Replace(outputStr, "ei", "ey", -1)
	outputStr = strings.Replace(outputStr, "e'i", "3y", -1)
	outputStr = strings.Replace(outputStr, "a'", "@", -1)
	outputStr = strings.Replace(outputStr, "e'", "3", -1)
	outputStr = strings.Replace(outputStr, "i'", "!", -1)
	outputStr = strings.Replace(outputStr, "o'", "%", -1)
	outputStr = strings.Replace(outputStr, "u'", "&", -1)
	outputStr = strings.Replace(outputStr, "gw", "$", -1)

	outputStr = strings.Replace(outputStr, "g", "k", -1)

	outputStr = strings.Replace(outputStr, "p'", "p*", -1)
	outputStr = strings.Replace(outputStr, "t'", "t*", -1)
	outputStr = strings.Replace(outputStr, "k'", "k*", -1)

	for charIndex, character := range outputStr {
		if IsSonorant(string(character)) {
			if charIndex != 0 {
				if IsConsonant(string(outputStr[charIndex-1])) || IsSonorant(string(outputStr[charIndex-1])) && (!IsSemivowel(string(outputStr[charIndex-1]))) {
					if string(character) == "m" {
						outputStr = string(outputStr[:charIndex]) + "8" + outputStr[charIndex+1:]
					} else if string(character) == "n" {
						outputStr = string(outputStr[:charIndex]) + "9" + outputStr[charIndex+1:]
					} else if string(character) == "l" {
						outputStr = string(outputStr[:charIndex]) + "0" + outputStr[charIndex+1:]
					}
				}
			} else {
				if IsConsonant(string(outputStr[charIndex+1])) || IsSonorant(string(outputStr[charIndex+1])) {
					if string(character) == "l" {
						outputStr = string(outputStr[:charIndex]) + "6" + outputStr[charIndex+1:]
					} else if string(character) == "n" {
						outputStr = string(outputStr[:charIndex]) + "7" + outputStr[charIndex+1:]
					}
				}
			}
		} else if IsAllophonicallyVoiced(string(character)) {
			if charIndex == 0 {
				if !(IsConsonant(string(outputStr[charIndex+1]))) {
					if string(character) == "t" {
						outputStr = "d" + outputStr[charIndex+1:]
					} else if string(character) == "p" {
						outputStr = "b" + outputStr[charIndex+1:]
					} else if string(character) == "k" {
						outputStr = "g" + outputStr[charIndex+1:]
					} else if string(character) == "$" {
						outputStr = "#" + outputStr[charIndex+1:]
					}
				} else if IsConsonant(string(outputStr[charIndex+1])) && string(character) == "j" {
					outputStr = "c" + outputStr[charIndex+1:]
				}
			} else if charIndex != len(outputStr)-1 && string(outputStr[charIndex+1]) != " " {
				if !(IsConsonant(string(outputStr[charIndex-1]))) && !(IsConsonant(string(outputStr[charIndex+1]))) {
					if string(character) == "t" {
						outputStr = string(outputStr[:charIndex]) + "d" + outputStr[charIndex+1:]
					} else if string(character) == "p" {
						outputStr = string(outputStr[:charIndex]) + "b" + outputStr[charIndex+1:]
					} else if string(character) == "k" {
						outputStr = string(outputStr[:charIndex]) + "g" + outputStr[charIndex+1:]
					} else if string(character) == "$" {
						outputStr = string(outputStr[:charIndex]) + "#" + outputStr[charIndex+1:]
					}
				} else if (IsConsonant(string(outputStr[charIndex-1])) || IsConsonant(string(outputStr[charIndex+1]))) && string(character) == "j" {
					outputStr = string(outputStr[:charIndex]) + "c" + outputStr[charIndex+1:]
				}
			} else if charIndex == len(outputStr)-1 || string(outputStr[charIndex+1]) == " " {
				if string(character) == "j" {
					outputStr = outputStr[:charIndex] + "c"
				}
			}
		}
	}

	if IsConsonant(string((outputStr[0]))) && IsConsonant(string((outputStr[1]))) {
		outputStr = "*" + outputStr
	}

	// sonorants after syllabic word-initial sonorants do not need to be recognized as such
	outputStr = strings.Replace(outputStr, "68", "6m", -1)
	outputStr = strings.Replace(outputStr, "69", "6n", -1)
	outputStr = strings.Replace(outputStr, "60", "6l", -1)
	outputStr = strings.Replace(outputStr, "78", "7m", -1)
	outputStr = strings.Replace(outputStr, "79", "7n", -1)
	outputStr = strings.Replace(outputStr, "70", "7l", -1)

	return outputStr
}

func normalizePacifique(inputStr string) string {
	outputStr := inputStr
	outputStr = strings.Replace(outputStr, "ai", "ay", -1)
	outputStr = strings.Replace(outputStr, "ei", "ey", -1)
	outputStr = strings.Replace(outputStr, "goa", "$a", -1)
	outputStr = strings.Replace(outputStr, "goe", "$e", -1)
	outputStr = strings.Replace(outputStr, "goi", "$i", -1)
	outputStr = strings.Replace(outputStr, "go", "$", -1)

	outputStr = strings.Replace(outputStr, "o", "u", -1)
	outputStr = strings.Replace(outputStr, "ô", "o", -1)

	outputStr = strings.Replace(outputStr, "g", "k", -1)
	outputStr = strings.Replace(outputStr, "tj", "c", -1)

	for charIndex, character := range outputStr {
		if IsAllophonicallyVoiced(string(character)) {
			if charIndex == 0 {
				if !(IsConsonant(string(outputStr[charIndex+1]))) {
					if string(character) == "t" {
						outputStr = "d" + outputStr[charIndex+1:]
					} else if string(character) == "p" {
						outputStr = "b" + outputStr[charIndex+1:]
					} else if string(character) == "k" {
						outputStr = "g" + outputStr[charIndex+1:]
					} else if string(character) == "c" {
						outputStr = "j" + outputStr[charIndex+1:]
					} else if string(character) == "$" {
						outputStr = "#" + outputStr[charIndex+1:]
					}
				}
			} else if charIndex != len(outputStr)-1 {
				if !(IsConsonant(string(outputStr[charIndex-1]))) && !(IsConsonant(string(outputStr[charIndex+1]))) {
					if string(character) == "t" {
						outputStr = string(outputStr[:charIndex]) + "d" + outputStr[charIndex+1:]
					} else if string(character) == "p" {
						outputStr = string(outputStr[:charIndex]) + "b" + outputStr[charIndex+1:]
					} else if string(character) == "k" {
						outputStr = string(outputStr[:charIndex]) + "g" + outputStr[charIndex+1:]
					} else if string(character) == "c" {
						outputStr = string(outputStr[:charIndex]) + "j" + outputStr[charIndex+1:]
					} else if string(character) == "$" {
						outputStr = string(outputStr[:charIndex]) + "#" + outputStr[charIndex+1:]
					}
				}
			}
		}
	}

	// sonorants after syllabic word-initial sonorants do not need to be recognized as such
	outputStr = strings.Replace(outputStr, "68", "6m", -1)
	outputStr = strings.Replace(outputStr, "69", "6n", -1)
	outputStr = strings.Replace(outputStr, "60", "6l", -1)
	outputStr = strings.Replace(outputStr, "78", "7m", -1)
	outputStr = strings.Replace(outputStr, "79", "7n", -1)
	outputStr = strings.Replace(outputStr, "70", "7l", -1)

	return outputStr
}

func normalizeRand(inputStr string) string {
	outputStr := inputStr

	outputStr = strings.Replace(outputStr, "a-", "ā", -1)
	outputStr = strings.Replace(outputStr, "a/", "ă", -1)
	outputStr = strings.Replace(outputStr, "a!", "â", -1)
	outputStr = strings.Replace(outputStr, "e/", "ĕ", -1)
	outputStr = strings.Replace(outputStr, "i/", "ĭ", -1)
	outputStr = strings.Replace(outputStr, "o/", "ŏ", -1)
	outputStr = strings.Replace(outputStr, "u/", "ŭ", -1)
	outputStr = strings.Replace(outputStr, "tc", "tç", -1)
	outputStr = strings.Replace(outputStr, "'", "", -1)

	if string([]rune(outputStr)[0:2]) == "ŭl" {
		outputStr = strings.Replace(outputStr, "ŭl", "6", -1)
	} else if string([]rune(outputStr)[0:2]) == "ŭn" {
		outputStr = strings.Replace(outputStr, "ŭn", "7", -1)
	}

	outputStr = strings.Replace(outputStr, "ŭm", "8", -1)
	outputStr = strings.Replace(outputStr, "ŭn", "9", -1)
	outputStr = strings.Replace(outputStr, "ŭl", "0", -1)

	outputStr = strings.Replace(outputStr, "ch", "c", -1)

	outputStr = strings.Replace(outputStr, "āā", "3y", -1)
	outputStr = strings.Replace(outputStr, "ā", "3", -1)
	outputStr = strings.Replace(outputStr, "ŭŭ", "*", -1)
	outputStr = strings.Replace(outputStr, "ŭ", "*", -1)
	outputStr = strings.Replace(outputStr, "ăă", "@", -1)
	outputStr = strings.Replace(outputStr, "ă", "a", -1)
	outputStr = strings.Replace(outputStr, "â", "@", -1)
	outputStr = strings.Replace(outputStr, "aa", "@", -1)
	outputStr = strings.Replace(outputStr, "eei", "@y", -1)
	outputStr = strings.Replace(outputStr, "ei", "ay", -1)
	outputStr = strings.Replace(outputStr, "oow", "@w", -1)
	outputStr = strings.Replace(outputStr, "ee", "!", -1)
	outputStr = strings.Replace(outputStr, "e", "!", -1)
	outputStr = strings.Replace(outputStr, "ĕĕ", "3", -1)
	outputStr = strings.Replace(outputStr, "ĕ", "e", -1)
	outputStr = strings.Replace(outputStr, "ĭĭ", "!", -1)
	outputStr = strings.Replace(outputStr, "ĭ", "i", -1)
	outputStr = strings.Replace(outputStr, "uu", "&", -1)
	outputStr = strings.Replace(outputStr, "u", "&", -1)
	outputStr = strings.Replace(outputStr, "ow", "aw", -1)
	outputStr = strings.Replace(outputStr, "oo", "&", -1)
	outputStr = strings.Replace(outputStr, "o", "%", -1)
	outputStr = strings.Replace(outputStr, "ŏŏ", "u", -1)
	outputStr = strings.Replace(outputStr, "ŏ", "o", -1)
	outputStr = strings.Replace(outputStr, "h", "q", -1)
	outputStr = strings.Replace(outputStr, "tç", "c", -1)
	outputStr = strings.Replace(outputStr, "dj", "j", -1)
	outputStr = strings.Replace(outputStr, "gw", "#", -1)
	outputStr = strings.Replace(outputStr, "kw", "$", -1)

	// sonorants after syllabic word-initial sonorants do not need to be recognized as such
	outputStr = strings.Replace(outputStr, "68", "6m", -1)
	outputStr = strings.Replace(outputStr, "69", "6n", -1)
	outputStr = strings.Replace(outputStr, "60", "6l", -1)
	outputStr = strings.Replace(outputStr, "78", "7m", -1)
	outputStr = strings.Replace(outputStr, "79", "7n", -1)
	outputStr = strings.Replace(outputStr, "70", "7l", -1)

	return outputStr
}

func normalizeLexicon(inputStr string) string {
	outputStr := inputStr
	outputStr = strings.Replace(outputStr, ":", "'", -1)
	outputStr = normalizeFrancisSmith(outputStr)
	return outputStr
}

func normalizeMetallic(inputStr string) string {
	outputStr := inputStr
	outputStr = strings.Replace(outputStr, "e!", "ê", -1)

	if string([]rune(outputStr)[0:2]) == "êl" {
		outputStr = strings.Replace(outputStr, "êl", "6", -1)
	} else if string([]rune(outputStr)[0:2]) == "ên" {
		outputStr = strings.Replace(outputStr, "ên", "7", -1)
	}

	outputStr = strings.Replace(outputStr, "êm", "8", -1)
	outputStr = strings.Replace(outputStr, "ên", "9", -1)
	outputStr = strings.Replace(outputStr, "êl", "0", -1)
	outputStr = strings.Replace(outputStr, "ê", "*", -1)

	outputStr = strings.Replace(outputStr, "ch", "c", -1)
	outputStr = strings.Replace(outputStr, "kw", "$", -1)
	outputStr = strings.Replace(outputStr, "gw", "#", -1)

	outputStr = strings.Replace(outputStr, "à", "@", -1)
	outputStr = strings.Replace(outputStr, "è", "3", -1)
	outputStr = strings.Replace(outputStr, "ì", "!", -1)
	outputStr = strings.Replace(outputStr, "ò", "%", -1)
	outputStr = strings.Replace(outputStr, "ù", "&", -1)

	// sonorants after syllabic word-initial sonorants do not need to be recognized as such
	outputStr = strings.Replace(outputStr, "68", "6m", -1)
	outputStr = strings.Replace(outputStr, "69", "6n", -1)
	outputStr = strings.Replace(outputStr, "60", "6l", -1)
	outputStr = strings.Replace(outputStr, "78", "7m", -1)
	outputStr = strings.Replace(outputStr, "79", "7n", -1)
	outputStr = strings.Replace(outputStr, "70", "7l", -1)

	return outputStr
}

func encodeOutput(inputStr string) Output {
	var OutputWords Output
	OutputWords.FrancisSmith = inputStr
	OutputWords.Listuguj = inputStr
	OutputWords.Pacifique = inputStr
	OutputWords.Rand = inputStr
	OutputWords.Lexicon = inputStr
	OutputWords.Metallic = inputStr

	// francis-smith
	if string(OutputWords.FrancisSmith[0]) == "*" {
		OutputWords.FrancisSmith = OutputWords.FrancisSmith[1:]
	}
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "*", "ɨ", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "@", "a'", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "3", "e'", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "!", "i'", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "%", "o'", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "&", "u'", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "c", "j", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "d", "t", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "b", "p", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "g", "k", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "6", "l'", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "7", "n'", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "8", "m", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "9", "n", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "0", "l", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "#", "kw", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "$", "kw", -1)

	// listuguj
	if string(OutputWords.Listuguj[0]) == "*" {
		OutputWords.Listuguj = OutputWords.Listuguj[1:]
	}
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "*", "'", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "@", "a'", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "3", "e'", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "!", "i'", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "%", "o'", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "&", "u'", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "c", "j", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "d", "t", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "b", "p", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "k", "g", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "$", "gw", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "#", "gw", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "6", "l", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "7", "n", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "8", "m", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "9", "n", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "0", "l", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "y", "i", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "eii", "e'i", -1) // e.g. weleyi > wele'i
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "ii", "i", -1)    // have to replace double i created by previous line
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "-", "", -1)

	//pacifique
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "cc", "c", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "*", "e", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "o", "ô", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "u", "o", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "@", "a", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "3", "e", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "!", "i", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "%", "ô", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "&", "o", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "j", "tj", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "d", "t", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "b", "p", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "k", "g", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "6", "el", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "7", "en", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "8", "m", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "9", "n", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "0", "l", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "c", "tj", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "y", "i", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "ii", "i", -1) // have to replace double i created by previous line
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "w", "o", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "q", "g", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "#", "go", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "$", "go", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "oo", "o", -1) // have to replace double o created by previous lines (i.e. -wo-)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "-", "", -1)

	// rand? more work needed for sure
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "cc", "c", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "ey", "ā", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "3y", "āā", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "*", "ŭ", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "a", "ă", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "@", "â", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "@y", "eei", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "@w", "oow", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "e", "ĕ", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "ăy", "ei", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "3", "ā", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "i", "ĭ", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "!", "e", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "o", "ŏ", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "ăw", "ow", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "%", "o", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "u", "ŏŏ", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "&", "oo", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "q", "h", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "c", "ch", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "j", "dj", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "6", "ŭl", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "7", "ŭn", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "8", "ŭm", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "99", "9n", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "9", "ŭn", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "0", "ŭl", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "#", "gw", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "$", "kw", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "-", "", -1)

	// lexicon
	OutputWords.Lexicon = OutputWords.FrancisSmith
	OutputWords.Lexicon = strings.Replace(OutputWords.Lexicon, "'", ":", -1)
	OutputWords.Lexicon = strings.Replace(OutputWords.Lexicon, "l:", "l'", -1) // have to reconvert "l:" etc. probably easier to do this than to be smarter about the previous line
	OutputWords.Lexicon = strings.Replace(OutputWords.Lexicon, "n:", "n'", -1)

	// metallic
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "cc", "c", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "*", "ê", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "@", "à", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "3", "è", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "!", "ì", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "%", "ò", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "&", "ù", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "c", "ch", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "6", "êl", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "7", "ên", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "8", "êm", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "99", "9n", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "9", "ên", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "0", "êl", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "#", "gw", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "$", "kw", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "-", "", -1)

	return OutputWords
}
