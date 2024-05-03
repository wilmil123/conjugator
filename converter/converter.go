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
// "=" is for /xw/ (coarticulated)
// "8" is syllabic /m/, non-initial
// "9" is syllabic /n/, non-initial
// "0" is syllabic /l/, non-initial
// "6" is initial syllabic /l/
// "7" is initial syllabic /n/
// "+" is initial syllabic /m/
// "j" is voiced /dʒ/
// "c" is voiceless /tʃ/
// many unique sequences are used for rand orthography special characters, in case the user cannot type them. these can be found in the character substitution table

// todo
// put more work into rand orthography
// add a system of sequence escaping

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

func retainInitialCapital(OutputWords Output) Output {
	OutputWords.FrancisSmith = fmt.Sprintf("%s%s", strings.ToUpper(string([]rune(OutputWords.FrancisSmith)[0])), string([]rune(OutputWords.FrancisSmith)[1:]))
	OutputWords.Listuguj = fmt.Sprintf("%s%s", strings.ToUpper(string([]rune(OutputWords.Listuguj)[0])), string([]rune(OutputWords.Listuguj)[1:]))
	OutputWords.Pacifique = fmt.Sprintf("%s%s", strings.ToUpper(string([]rune(OutputWords.Pacifique)[0])), string([]rune(OutputWords.Pacifique)[1:]))
	OutputWords.Rand = fmt.Sprintf("%s%s", strings.ToUpper(string([]rune(OutputWords.Rand)[0])), string([]rune(OutputWords.Rand)[1:]))
	OutputWords.Lexicon = fmt.Sprintf("%s%s", strings.ToUpper(string([]rune(OutputWords.Lexicon)[0])), string([]rune(OutputWords.Lexicon)[1:]))
	OutputWords.Metallic = fmt.Sprintf("%s%s", strings.ToUpper(string([]rune(OutputWords.Metallic)[0])), string([]rune(OutputWords.Metallic)[1:]))
	return OutputWords
}

func escapeSequences(InputStr string) ([]string, []string) { // handling escaped sequences
	var strToConvert []string  // storing those to be converted
	var strToConserve []string // storing those not to be converted
	localStr := InputStr
	for escapedSequences := 0; escapedSequences <= strings.Count(InputStr, "{"); escapedSequences++ {
		splitStr1 := []string{"", ""}
		if strings.Index(localStr, "{") != len(localStr)-1 && strings.Contains(localStr, "{") {
			splitStr1[0], splitStr1[1], _ = strings.Cut(localStr, "{")
			strToConvert = append(strToConvert, splitStr1[0])
		} else {
			splitStr1[0], _ = strings.CutSuffix(localStr, "{")
			strToConvert = append(strToConvert, splitStr1[0])
			break
		}
		splitStr2 := []string{"", ""}
		if strings.Index(splitStr1[1], "}") != len(splitStr1[1])-1 && strings.Contains(splitStr1[1], "}") {
			splitStr2[0], splitStr2[1], _ = strings.Cut(splitStr1[1], "}")
			strToConserve = append(strToConserve, splitStr2[0])
		} else {
			splitStr2[0], _ = strings.CutSuffix(splitStr1[1], "}")
			strToConserve = append(strToConserve, splitStr2[0])
			break
		}
		localStr = splitStr2[1]
	}
	return strToConvert, strToConserve
}

func collapseStrings(OutputWords Output, normalStr []string, strToConserve []string, conservedSequenceInitial bool) Output {
	var localFrancisSmith []string
	var localListuguj []string
	var localPacifique []string
	var localRand []string
	var localLexicon []string
	var localMetallic []string
	if len(normalStr) > len(strToConserve) {
		for strCount := range normalStr {
			var localOutput Output = encodeOutput(normalStr[strCount])
			if !conservedSequenceInitial {
				localFrancisSmith = append(localFrancisSmith, localOutput.FrancisSmith)
				localListuguj = append(localListuguj, localOutput.Listuguj)
				localPacifique = append(localPacifique, localOutput.Pacifique)
				localRand = append(localRand, localOutput.Rand)
				localLexicon = append(localLexicon, localOutput.Lexicon)
				localMetallic = append(localMetallic, localOutput.Metallic)

				if strCount < len(strToConserve) {
					localFrancisSmith = append(localFrancisSmith, strToConserve[strCount])
					localListuguj = append(localListuguj, strToConserve[strCount])
					localPacifique = append(localPacifique, strToConserve[strCount])
					localRand = append(localRand, strToConserve[strCount])
					localLexicon = append(localLexicon, strToConserve[strCount])
					localMetallic = append(localMetallic, strToConserve[strCount])
				}
			} else {
				if strCount < len(strToConserve) {
					localFrancisSmith = append(localFrancisSmith, strToConserve[strCount])
					localListuguj = append(localListuguj, strToConserve[strCount])
					localPacifique = append(localPacifique, strToConserve[strCount])
					localRand = append(localRand, strToConserve[strCount])
					localLexicon = append(localLexicon, strToConserve[strCount])
					localMetallic = append(localMetallic, strToConserve[strCount])
				}

				localFrancisSmith = append(localFrancisSmith, localOutput.FrancisSmith)
				localListuguj = append(localListuguj, localOutput.Listuguj)
				localPacifique = append(localPacifique, localOutput.Pacifique)
				localRand = append(localRand, localOutput.Rand)
				localLexicon = append(localLexicon, localOutput.Lexicon)
				localMetallic = append(localMetallic, localOutput.Metallic)
			}
		}
	} else {
		for strCount := range strToConserve {
			var localOutput Output = encodeOutput(normalStr[strCount])
			if !conservedSequenceInitial {
				if strCount < len(normalStr) {
					localFrancisSmith = append(localFrancisSmith, localOutput.FrancisSmith)
					localListuguj = append(localListuguj, localOutput.Listuguj)
					localPacifique = append(localPacifique, localOutput.Pacifique)
					localRand = append(localRand, localOutput.Rand)
					localLexicon = append(localLexicon, localOutput.Lexicon)
					localMetallic = append(localMetallic, localOutput.Metallic)
				}

				localFrancisSmith = append(localFrancisSmith, strToConserve[strCount])
				localListuguj = append(localListuguj, strToConserve[strCount])
				localPacifique = append(localPacifique, strToConserve[strCount])
				localRand = append(localRand, strToConserve[strCount])
				localLexicon = append(localLexicon, strToConserve[strCount])
				localMetallic = append(localMetallic, strToConserve[strCount])
			} else {
				localFrancisSmith = append(localFrancisSmith, strToConserve[strCount])
				localListuguj = append(localListuguj, strToConserve[strCount])
				localPacifique = append(localPacifique, strToConserve[strCount])
				localRand = append(localRand, strToConserve[strCount])
				localLexicon = append(localLexicon, strToConserve[strCount])
				localMetallic = append(localMetallic, strToConserve[strCount])

				if strCount < len(normalStr) {
					localFrancisSmith = append(localFrancisSmith, localOutput.FrancisSmith)
					localListuguj = append(localListuguj, localOutput.Listuguj)
					localPacifique = append(localPacifique, localOutput.Pacifique)
					localRand = append(localRand, localOutput.Rand)
					localLexicon = append(localLexicon, localOutput.Lexicon)
					localMetallic = append(localMetallic, localOutput.Metallic)
				}
			}
		}
	}
	OutputWords.FrancisSmith = strings.Join(localFrancisSmith, "")
	OutputWords.Listuguj = strings.Join(localListuguj, "")
	OutputWords.Pacifique = strings.Join(localPacifique, "")
	OutputWords.Rand = strings.Join(localRand, "")
	OutputWords.Lexicon = strings.Join(localLexicon, "")
	OutputWords.Metallic = strings.Join(localMetallic, "")

	return OutputWords
}

func orthoIndexHandler(writer http.ResponseWriter, reader *http.Request) {
	var normalStr []string // the "normal" string is that which is normalized for conversion to any orthography
	var wasUpper bool
	var strToConvert []string         // storing those to be converted
	var strToConserve []string        // storing those not to be converted
	var conservedSequenceInitial bool // whether the initial is a conserved sequence or not
	var PacifiqueDisclaimer bool = false
	var RandDisclaimer bool = false
	if reader.Method == http.MethodPost { // if the "go" button is pressed
		InputStr := reader.FormValue("wordinput")              // get the input string
		orthographyChoice := reader.FormValue("orthographies") // a string value correstponding to the orthography chosen by the user
		if InputStr != "" {                                    // if the input is not empty
			if strings.ToUpper(string([]rune(InputStr)[0])) == string([]rune(InputStr)[0]) { // if the first character's capital letter is equal to its value, i.e. it is a capital
				wasUpper = true // record that the first letter was a capital. in the future, could maybe try for multiple capitals?
			}
			InputStr = strings.ToLower(InputStr)
			if strings.Contains(InputStr, "{") && strings.Contains(InputStr, "}") {
				if strings.Index(InputStr, "{") == 0 {
					conservedSequenceInitial = true
				} else {
					conservedSequenceInitial = false
				}
				strToConvert, strToConserve = escapeSequences(InputStr)
			} else {
				strToConvert = append(strToConvert, InputStr)
			}
			switch orthographyChoice {
			case "francissmith":
				for strCount := range strToConvert {
					normalStr = append(normalStr, normalizeFrancisSmith(strToConvert[strCount]))
				}
			case "listuguj":
				for strCount := range strToConvert {
					normalStr = append(normalStr, normalizeListuguj(strToConvert[strCount]))
				}
			case "pacifique":
				for strCount := range strToConvert {
					normalStr = append(normalStr, normalizePacifique(strToConvert[strCount]))
				}
				PacifiqueDisclaimer = true
			case "rand":
				for strCount := range strToConvert {
					normalStr = append(normalStr, normalizeRand(strToConvert[strCount]))
				}
				RandDisclaimer = true
			case "lexicon":
				for strCount := range strToConvert {
					normalStr = append(normalStr, normalizeLexicon(strToConvert[strCount]))
				}
			case "metallic":
				for strCount := range strToConvert {
					normalStr = append(normalStr, normalizeMetallic(strToConvert[strCount]))
				}
			default:
				fmt.Println("orthography type missing")
			}
		}
	} else { // if the button was not pressed (i.e. on first load of the page without cache)
		normalStr = append(normalStr, normalizeFrancisSmith("put*p")) // default is "put*p"
	}

	var OutputWords Output
	OutputWords = collapseStrings(OutputWords, normalStr, strToConserve, conservedSequenceInitial)
	OutputWords.PacifiqueDisclaimer = PacifiqueDisclaimer
	OutputWords.RandDisclaimer = RandDisclaimer
	if wasUpper && len(normalStr) > 1 {
		OutputWords = retainInitialCapital(OutputWords)
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
		"#",
		"$",
		"p",
		"b",
		"q",
		"=",
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

func IsDelineator(category string) bool { // returns true if the passed slice is in this list
	switch category {
	case
		" ",
		".",
		",":
		return true
	}
	return false
}

func IsAllophonicallyVoiced(category string) bool { // returns true if the passed slice is in this list
	switch category {
	case
		"c",
		"k",
		"$",
		"p",
		"t":
		return true
	}
	return false
}

func IsLowBackVowel(category string) bool { // returns true if the passed slice is in this list
	switch category {
	case
		"a",
		"@",
		"o",
		"%":
		return true
	}
	return false
}

// sonorants after syllabic word-initial sonorants do not need to be recognized as such
func fixSonorantDistribution(outputStr string) string {
	outputStr = strings.Replace(outputStr, "68", "6m", -1)
	outputStr = strings.Replace(outputStr, "69", "6n", -1)
	outputStr = strings.Replace(outputStr, "60", "6l", -1)
	outputStr = strings.Replace(outputStr, "78", "7m", -1)
	outputStr = strings.Replace(outputStr, "79", "7n", -1)
	outputStr = strings.Replace(outputStr, "70", "7l", -1)

	outputStr = strings.Replace(outputStr, "99", "9n", -1)
	outputStr = strings.Replace(outputStr, "n9", "nn", -1)
	return outputStr
}

// pacifique and rand are inconsistent with their renderings of the uvular/velar fricative; this function attempts to resolve some of that
func resolveUvularFricative(outputStr string) string {
	for charIndex, character := range outputStr {
		if string(character) == "g" || string(character) == "k" {
			if charIndex > 0 && charIndex != len(outputStr)-1 {
				if IsLowBackVowel(string(outputStr[charIndex-1])) &&
					(string(outputStr[charIndex+1]) != "i" && string(outputStr[charIndex+1]) != "!") {
					outputStr = fmt.Sprintf("%sq%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
				}
			} else if charIndex == len(outputStr)-1 {
				if IsLowBackVowel(string(outputStr[charIndex-1])) {
					outputStr = fmt.Sprintf("%sq", string(outputStr[:charIndex]))
				}
			}
		} else if string(character) == "#" || string(character) == "$" {
			if charIndex == 0 && IsLowBackVowel(string(outputStr[charIndex+1])) {
				outputStr = fmt.Sprintf("=%s", string(outputStr[1:]))
			} else if charIndex != len(outputStr)-1 {
				if IsLowBackVowel(string(outputStr[charIndex-1])) &&
					(string(outputStr[charIndex+1]) != "i" && string(outputStr[charIndex+1]) != "!") {
					outputStr = fmt.Sprintf("%s=%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
				}
			} else if charIndex == len(outputStr)-1 {
				if IsLowBackVowel(string(outputStr[charIndex-1])) {
					outputStr = fmt.Sprintf("%s=", string(outputStr[:charIndex]))
				}
			}
		}
	}
	return outputStr
}

// turns francis-smith into unified orthography
func normalizeFrancisSmith(inputStr string) string {
	outputStr := inputStr
	// below are standard character replacements to 1 glyph 'unified orthography' values
	outputStr = strings.Replace(outputStr, "ɨ", "*", -1)
	outputStr = strings.Replace(outputStr, "a'", "@", -1)
	outputStr = strings.Replace(outputStr, "e'", "3", -1)
	outputStr = strings.Replace(outputStr, "i'", "!", -1)
	outputStr = strings.Replace(outputStr, "o'", "%", -1)
	outputStr = strings.Replace(outputStr, "u'", "&", -1)
	outputStr = strings.Replace(outputStr, "kw", "$", -1)
	outputStr = strings.Replace(outputStr, "qw", "=", -1)
	outputStr = strings.Replace(outputStr, "j", "c", -1) // replace with voiceless allophone for consistency with p, t, k

	// loop through every character in the input string
	for charIndex, character := range outputStr {
		// if the current character is a sonorant
		if IsSonorant(string(character)) {
			// if the sonorant is not initial, and the preceding character is a consonant or a sonorant, but not a semivowel
			if charIndex != 0 && (IsConsonant(string(outputStr[charIndex-1])) || IsSonorant(string(outputStr[charIndex-1]))) &&
				(!IsSemivowel(string(outputStr[charIndex-1]))) {
				// replace these with their allophonic variants
				if string(character) == "m" {
					outputStr = fmt.Sprintf("%s8%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
				} else if string(character) == "n" {
					outputStr = fmt.Sprintf("%s9%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
				} else if string(character) == "l" {
					outputStr = fmt.Sprintf("%s0%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
				}
			}
		} else if IsAllophonicallyVoiced(string(character)) { // if the current character is allophonically voiced
			if charIndex == 0 && !(IsConsonant(string(outputStr[charIndex+1]))) { // if the consonant is at the beginning of a word and the next character is not a consonant
				if string(character) == "t" {
					outputStr = fmt.Sprintf("d%s", string(outputStr[charIndex+1:]))
				} else if string(character) == "p" {
					outputStr = fmt.Sprintf("b%s", string(outputStr[charIndex+1:]))
				} else if string(character) == "k" {
					outputStr = fmt.Sprintf("g%s", string(outputStr[charIndex+1:]))
				} else if string(character) == "$" {
					outputStr = fmt.Sprintf("#%s", string(outputStr[charIndex+1:]))
				} else if string(character) == "c" {
					outputStr = fmt.Sprintf("j%s", string(outputStr[charIndex+1:]))
				}
			} else if charIndex != len(outputStr)-1 && !IsDelineator((string(outputStr[charIndex+1]))) {
				if !(IsConsonant(string(outputStr[charIndex-1]))) && !(IsConsonant(string(outputStr[charIndex+1]))) {
					// if the consonant is word-final and is not in a cluster, replace with voiced variants
					if string(character) == "t" {
						outputStr = fmt.Sprintf("%sd%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "p" {
						outputStr = fmt.Sprintf("%sb%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "k" {
						outputStr = fmt.Sprintf("%sg%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "c" {
						outputStr = fmt.Sprintf("%sj%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "$" {
						outputStr = fmt.Sprintf("%s#%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					}
				}
			}
		}
	}

	// if the first two characters are consonants, begin the word with a schwa
	if IsConsonant(string((outputStr[0]))) && IsConsonant(string((outputStr[1]))) {
		outputStr = fmt.Sprintf("*%s", outputStr)
	}

	// initial syllabic consonants are rendered differently in francis-smith & lexicon, so must be recognized here
	outputStr = strings.Replace(outputStr, "l'", "6", -1)
	outputStr = strings.Replace(outputStr, "n'", "7", -1)
	outputStr = strings.Replace(outputStr, "m'", "+", -1)

	// sonorants after syllabic word-initial sonorants do not need to be recognized as such
	outputStr = fixSonorantDistribution(outputStr)

	return outputStr
}

// turns listuguj into unified orthography
func normalizeListuguj(inputStr string) string {
	outputStr := inputStr

	// listuguj does not recognize /j/ as a semivowel, but it is easy to replace these since /j/ only appears after vowels
	outputStr = strings.Replace(outputStr, "ai", "ay", -1)
	outputStr = strings.Replace(outputStr, "a'i", "@y", -1)
	outputStr = strings.Replace(outputStr, "ei", "ey", -1)
	outputStr = strings.Replace(outputStr, "e'i", "3y", -1)

	// standard 1 glyph 'unified orthography' character conversions
	outputStr = strings.Replace(outputStr, "a'", "@", -1)
	outputStr = strings.Replace(outputStr, "e'", "3", -1)
	outputStr = strings.Replace(outputStr, "i'", "!", -1)
	outputStr = strings.Replace(outputStr, "o'", "%", -1)
	outputStr = strings.Replace(outputStr, "u'", "&", -1)

	// replace gw, g, j with voiceless variants
	outputStr = strings.Replace(outputStr, "gw", "$", -1)
	outputStr = strings.Replace(outputStr, "qw", "=", -1)
	outputStr = strings.Replace(outputStr, "g", "k", -1)
	outputStr = strings.Replace(outputStr, "j", "c", -1)

	// listuguj uses the apostrophe for both schwa and vowel length, but it is easy to find the schwas since they will follow consonants
	outputStr = strings.Replace(outputStr, "p'", "p*", -1)
	outputStr = strings.Replace(outputStr, "t'", "t*", -1)
	outputStr = strings.Replace(outputStr, "k'", "k*", -1)
	outputStr = strings.Replace(outputStr, "s'", "s*", -1)
	outputStr = strings.Replace(outputStr, "c'", "c*", -1)
	outputStr = strings.Replace(outputStr, "q'", "q*", -1)

	// loop through every character in the string
	for charIndex, character := range outputStr {
		// if the character is a sonorant
		if IsSonorant(string(character)) {
			// if this sonorant is not word-initial
			if charIndex != 0 {
				// if the previous character is a consonant or sonorant but is not a semivowel, replace with syllabic variants
				if IsConsonant(string(outputStr[charIndex-1])) || IsSonorant(string(outputStr[charIndex-1])) && (!IsSemivowel(string(outputStr[charIndex-1]))) {
					if string(character) == "m" {
						outputStr = fmt.Sprintf("%s8%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "n" {
						outputStr = fmt.Sprintf("%s9%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "l" {
						outputStr = fmt.Sprintf("%s0%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					}
				} else if (IsConsonant(string(outputStr[charIndex+1])) || IsSonorant(string(outputStr[charIndex+1]))) && IsDelineator(string(outputStr[charIndex-1])) {
					// if the character is a sonorant that can be voiced word-initially and the previous character is a delineator
					if string(character) == "l" {
						outputStr = fmt.Sprintf("%s6%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "n" {
						outputStr = fmt.Sprintf("%s7%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "m" {
						outputStr = fmt.Sprintf("%s+%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					}
				}
			} else if charIndex == 0 { // else (if the sonorant is word initial)
				// if the following character is a consonant or sonorant, replace with word-initial syllabic variants
				if IsConsonant(string(outputStr[charIndex+1])) || IsSonorant(string(outputStr[charIndex+1])) {
					if string(character) == "l" {
						outputStr = fmt.Sprintf("%s6%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "n" {
						outputStr = fmt.Sprintf("%s7%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "m" {
						outputStr = fmt.Sprintf("%s+%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					}
				}
			}
		} else if IsAllophonicallyVoiced(string(character)) { // if this character is allophonically voiced
			if charIndex == 0 { // if this character is word-initial
				if !(IsConsonant(string(outputStr[charIndex+1]))) { // if the next character is not a consonant, replace with voiced variants
					if string(character) == "t" {
						outputStr = fmt.Sprintf("d%s", string(outputStr[charIndex+1:]))
					} else if string(character) == "p" {
						outputStr = fmt.Sprintf("b%s", string(outputStr[charIndex+1:]))
					} else if string(character) == "k" {
						outputStr = fmt.Sprintf("g%s", string(outputStr[charIndex+1:]))
					} else if string(character) == "$" {
						outputStr = fmt.Sprintf("#%s", string(outputStr[charIndex+1:]))
					} else if string(character) == "c" {
						outputStr = fmt.Sprintf("j%s", string(outputStr[charIndex+1:]))
					}
				}
			} else if charIndex != len(outputStr)-1 && !IsDelineator(string(outputStr[charIndex+1])) {
				// if this character is not the last, and the following character is not a delineator
				// if it is not surrounded by consonants, make it voiced
				if !(IsConsonant(string(outputStr[charIndex-1]))) && !(IsConsonant(string(outputStr[charIndex+1]))) {
					if string(character) == "t" {
						outputStr = fmt.Sprintf("%sd%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "p" {
						outputStr = fmt.Sprintf("%sb%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "k" {
						outputStr = fmt.Sprintf("%sg%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "$" {
						outputStr = fmt.Sprintf("%s#%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "c" {
						outputStr = fmt.Sprintf("%sj%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					}
				}
			}
		}
	}

	// if the first two characters are consonants, insert a schwa at the beginning
	if IsConsonant(string((outputStr[0]))) && IsConsonant(string((outputStr[1]))) {
		outputStr = fmt.Sprintf("*%s", outputStr)
	}

	// sonorants after syllabic word-initial sonorants do not need to be recognized as such
	outputStr = fixSonorantDistribution(outputStr)

	return outputStr
}

// turns pacifique into unified orthography (known problematic, but unlikely anything can be done)
func normalizePacifique(inputStr string) string {
	outputStr := inputStr

	// o! is ô from the character substitution table
	outputStr = strings.Replace(outputStr, "o!", "ô", -1)

	// replace i and o with /j/, /w/ when it is known they exist
	outputStr = strings.Replace(outputStr, "ai", "ay", -1)
	outputStr = strings.Replace(outputStr, "ao", "aw", -1)
	outputStr = strings.Replace(outputStr, "ei", "ey", -1)
	outputStr = strings.Replace(outputStr, "eo", "ew", -1)
	outputStr = strings.Replace(outputStr, "goa", "$a", -1)
	outputStr = strings.Replace(outputStr, "goe", "$e", -1)
	outputStr = strings.Replace(outputStr, "goi", "$i", -1)
	outputStr = strings.Replace(outputStr, "go", "$", -1)
	// this last replacement causes issues in front of other consonants, which have to be fixed
	outputStr = strings.Replace(outputStr, "$t", "got", -1)
	outputStr = strings.Replace(outputStr, "$p", "gop", -1)
	outputStr = strings.Replace(outputStr, "$g", "gog", -1)
	outputStr = strings.Replace(outputStr, "$$", "go$", -1)
	outputStr = strings.Replace(outputStr, "$l", "gol", -1)
	outputStr = strings.Replace(outputStr, "$m", "gom", -1)
	outputStr = strings.Replace(outputStr, "$n", "gon", -1)
	outputStr = strings.Replace(outputStr, "$w", "gow", -1)
	outputStr = strings.Replace(outputStr, "$y", "goy", -1)

	// pacifique uses ô for /o/, o for /u/
	outputStr = strings.Replace(outputStr, "o", "u", -1)
	outputStr = strings.Replace(outputStr, "ô", "o", -1)

	// replace double vowels (not necessarily indicated in pacifique)
	outputStr = strings.Replace(outputStr, "aa", "@", -1)
	outputStr = strings.Replace(outputStr, "ee", "3", -1)
	outputStr = strings.Replace(outputStr, "ii", "!", -1)
	outputStr = strings.Replace(outputStr, "oo", "%", -1)
	outputStr = strings.Replace(outputStr, "uu", "&", -1)

	// make every consonant voiceless for consistency
	outputStr = strings.Replace(outputStr, "g", "k", -1)
	outputStr = strings.Replace(outputStr, "tj", "c", -1)

	// loop over every character
	for charIndex, character := range outputStr {
		// if the character is allophonically voiced
		if IsAllophonicallyVoiced(string(character)) {
			// if it is the first character
			if charIndex == 0 {
				// if the next character is not a consonant, make it voiced
				if !(IsConsonant(string(outputStr[charIndex+1]))) {
					if string(character) == "t" {
						outputStr = fmt.Sprintf("d%s", string(outputStr[charIndex+1:]))
					} else if string(character) == "p" {
						outputStr = fmt.Sprintf("b%s", string(outputStr[charIndex+1:]))
					} else if string(character) == "k" {
						outputStr = fmt.Sprintf("g%s", string(outputStr[charIndex+1:]))
					} else if string(character) == "c" {
						outputStr = fmt.Sprintf("j%s", string(outputStr[charIndex+1:]))
					} else if string(character) == "$" {
						outputStr = fmt.Sprintf("#%s", string(outputStr[charIndex+1:]))
					}
				}
			} else if charIndex != len(outputStr)-1 && !IsDelineator(string(outputStr[charIndex+1])) { // if it is not the last character
				// if the previous and next characters are not consonants, make this consonant voiced
				if !(IsConsonant(string(outputStr[charIndex-1]))) && !(IsConsonant(string(outputStr[charIndex+1]))) {
					if string(character) == "t" {
						outputStr = fmt.Sprintf("%sd%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "p" {
						outputStr = fmt.Sprintf("%sb%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "k" {
						outputStr = fmt.Sprintf("%sg%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "c" {
						outputStr = fmt.Sprintf("%sj%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					} else if string(character) == "$" {
						outputStr = fmt.Sprintf("%s#%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
					}
				}
			}
		} else if string(character) == "u" {
			if charIndex == 0 || IsDelineator(string(outputStr[charIndex-1])) {
				if !IsConsonant(string(outputStr[charIndex+1])) && !IsSonorant(string(outputStr[charIndex+1])) && !IsSemivowel(string(outputStr[charIndex+1])) {
					outputStr = fmt.Sprintf("%sw%s", string(outputStr[:charIndex]), string(outputStr[charIndex+1:]))
				}
			}
		}
	}

	// sonorants after syllabic word-initial sonorants do not need to be recognized as such
	outputStr = fixSonorantDistribution(outputStr)

	// this function attempts to resolve some ambiguities with uvular fricatives in rand and pacifique
	outputStr = resolveUvularFricative(outputStr)

	return outputStr
}

// turns rand into unified orthography (known problematic, but can hopefully be fine-tuned)
func normalizeRand(inputStr string) string {
	outputStr := inputStr

	// per the character substitution table, replace these sequences with how they appear in rand orthography
	outputStr = strings.Replace(outputStr, "a-", "ā", -1)
	outputStr = strings.Replace(outputStr, "a/", "ă", -1)
	outputStr = strings.Replace(outputStr, "a!", "â", -1)
	outputStr = strings.Replace(outputStr, "a:", "ä", -1)
	outputStr = strings.Replace(outputStr, "e/", "ĕ", -1)
	outputStr = strings.Replace(outputStr, "e:", "ë", -1)
	outputStr = strings.Replace(outputStr, "i/", "ĭ", -1)
	outputStr = strings.Replace(outputStr, "i:", "ï", -1)
	outputStr = strings.Replace(outputStr, "o/", "ŏ", -1)
	outputStr = strings.Replace(outputStr, "o-", "ō", -1)
	outputStr = strings.Replace(outputStr, "o:", "ö", -1)
	outputStr = strings.Replace(outputStr, "u/", "ŭ", -1)
	outputStr = strings.Replace(outputStr, "u:", "ü", -1)
	outputStr = strings.Replace(outputStr, "tc", "tç", -1)

	// if the first two characters are ŭl/ŭn/ŭm or 'l/'n/'m (usage is inconsistent?)
	if string([]rune(outputStr)[0:2]) == "ŭl" || string([]rune(outputStr)[0:2]) == "'l" {
		outputStr = fmt.Sprintf("6%s", string([]rune(outputStr[2:])))
	} else if string([]rune(outputStr)[0:2]) == "ŭn" || string([]rune(outputStr)[0:2]) == "'n" {
		outputStr = fmt.Sprintf("7%s", string([]rune(outputStr[2:])))
	} else if string([]rune(outputStr)[0:2]) == "ŭm" || string([]rune(outputStr)[0:2]) == "'m" {
		outputStr = fmt.Sprintf("+%s", string([]rune(outputStr[2:])))
	}
	// replace with word-initial syllabic variants

	// an apostrophe in rand orthography marks stress, which is problematic in unified orthography since it is unpredictable and no other orthographies make use of it
	outputStr = strings.Replace(outputStr, "'", "", -1)

	// if ŭm, ŭn, ŭl are found elsewhere (i.e. in the middle of a word), replace them with non-word-initial syllabic variants
	outputStr = strings.Replace(outputStr, "ŭm", "8", -1)
	outputStr = strings.Replace(outputStr, "ŭn", "9", -1)
	outputStr = strings.Replace(outputStr, "ŭl", "0", -1)

	// for now, just replace the umlauted characters with regular ones. i think this is just meant to show vowel hiatus?
	outputStr = strings.Replace(outputStr, "ä", "a", -1)
	outputStr = strings.Replace(outputStr, "ë", "e", -1)
	outputStr = strings.Replace(outputStr, "ï", "ĭ", -1)
	outputStr = strings.Replace(outputStr, "ö", "o", -1)
	outputStr = strings.Replace(outputStr, "ü", "u", -1)

	// just a big list of character substitutions into unified orthography. could this be done better?
	outputStr = strings.Replace(outputStr, "eei", "@y", -1)
	outputStr = strings.Replace(outputStr, "oow", "@w", -1)
	outputStr = strings.Replace(outputStr, "ei", "ay", -1)
	outputStr = strings.Replace(outputStr, "ow", "aw", -1)
	outputStr = strings.Replace(outputStr, "āā", "3y", -1)
	outputStr = strings.Replace(outputStr, "ee", "!", -1)
	outputStr = strings.Replace(outputStr, "ŭŭ", "*", -1) // does this exist?
	outputStr = strings.Replace(outputStr, "ăă", "@", -1)
	outputStr = strings.Replace(outputStr, "aa", "@", -1)
	outputStr = strings.Replace(outputStr, "ĕĕ", "3", -1)
	outputStr = strings.Replace(outputStr, "ĭĭ", "!", -1) // does this exist?
	outputStr = strings.Replace(outputStr, "uu", "!w", -1)
	outputStr = strings.Replace(outputStr, "u", "iw", -1)
	outputStr = strings.Replace(outputStr, "oo", "u", -1)
	outputStr = strings.Replace(outputStr, "ŏŏ", "u", -1)

	outputStr = strings.Replace(outputStr, "ŭ", "*", -1)
	outputStr = strings.Replace(outputStr, "ă", "a", -1)
	outputStr = strings.Replace(outputStr, "â", "a", -1)
	outputStr = strings.Replace(outputStr, "e", "i", -1)
	outputStr = strings.Replace(outputStr, "ĕ", "e", -1)
	outputStr = strings.Replace(outputStr, "ā", "3", -1)
	if string([]rune(outputStr)[len([]rune(outputStr))-1]) == "3" {
		outputStr = fmt.Sprintf("%sey", string(outputStr[:(len(outputStr)-1)]))
	}
	outputStr = strings.Replace(outputStr, "ĭ", "i", -1)
	outputStr = strings.Replace(outputStr, "ō", "%", -1)
	outputStr = strings.Replace(outputStr, "ŏ", "o", -1)
	outputStr = strings.Replace(outputStr, "tç", "c", -1)
	// ch or tç are used in different versions of rand for /tʃ/
	outputStr = strings.Replace(outputStr, "ch", "c", -1)
	outputStr = strings.Replace(outputStr, "h", "q", -1)
	outputStr = strings.Replace(outputStr, "dj", "j", -1)
	outputStr = strings.Replace(outputStr, "gw", "#", -1)
	outputStr = strings.Replace(outputStr, "kw", "$", -1)

	// sonorants after syllabic word-initial sonorants do not need to be recognized as such
	outputStr = fixSonorantDistribution(outputStr)

	// this function attempts to resolve some ambiguities with uvular fricatives in rand and pacifique
	outputStr = resolveUvularFricative(outputStr)

	return outputStr
}

// turns lexicon into unified orthography
func normalizeLexicon(inputStr string) string {
	outputStr := inputStr

	// this is simple, since lexicon is essentially francis-smith with different vowel length indicators
	outputStr = strings.Replace(outputStr, ":", "'", -1)
	outputStr = normalizeFrancisSmith(outputStr)
	return outputStr
}

func normalizeMetallic(inputStr string) string {
	outputStr := inputStr

	// from the character substitution table, e! can be ê
	outputStr = strings.Replace(outputStr, "e!", "ê", -1)

	// if the first two characters are êl or ên
	if string([]rune(outputStr)[0:2]) == "êl" {
		outputStr = fmt.Sprintf("6%s", string([]rune(outputStr[2:])))
	} else if string([]rune(outputStr)[0:2]) == "ên" {
		outputStr = fmt.Sprintf("7%s", string([]rune(outputStr[2:])))
	} else if string([]rune(outputStr)[0:2]) == "êm" {
		outputStr = fmt.Sprintf("+%s", string([]rune(outputStr[2:])))
	}

	// replace remaining syllabic sonorants with their variants
	outputStr = strings.Replace(outputStr, "êm", "8", -1)
	outputStr = strings.Replace(outputStr, "ên", "9", -1)
	outputStr = strings.Replace(outputStr, "êl", "0", -1)
	// the schwa is ê in metallic, replace with ɨ
	outputStr = strings.Replace(outputStr, "ê", "*", -1)

	// since metallic makes voicing distinctions, no looping is required. just replace the voiceless variants with their 1-glyph counterparts when required
	outputStr = strings.Replace(outputStr, "ch", "c", -1)
	outputStr = strings.Replace(outputStr, "kw", "$", -1)
	outputStr = strings.Replace(outputStr, "gw", "#", -1)
	outputStr = strings.Replace(outputStr, "qw", "=", -1)

	// replace long vowels with their 1-glyph counterparts
	outputStr = strings.Replace(outputStr, "à", "@", -1)
	outputStr = strings.Replace(outputStr, "è", "3", -1)
	outputStr = strings.Replace(outputStr, "ì", "!", -1)
	outputStr = strings.Replace(outputStr, "ò", "%", -1)
	outputStr = strings.Replace(outputStr, "ù", "&", -1)

	// sonorants after syllabic word-initial sonorants do not need to be recognized as such
	outputStr = fixSonorantDistribution(outputStr)

	return outputStr
}

// takes unified orthography and turns it into the output for all different orthographies
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
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "+", "m'", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "8", "m", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "9", "n", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "0", "l", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "#", "kw", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "$", "kw", -1)
	OutputWords.FrancisSmith = strings.Replace(OutputWords.FrancisSmith, "=", "qw", -1)

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
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "=", "qw", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "6", "l", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "7", "n", -1)
	OutputWords.Listuguj = strings.Replace(OutputWords.Listuguj, "+", "m", -1)
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
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "+", "em", -1)
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
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "=", "go", -1)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "oo", "o", -1) // have to replace double o created by previous lines (i.e. -wo-)
	OutputWords.Pacifique = strings.Replace(OutputWords.Pacifique, "-", "", -1)

	// rand? more work needed for sure
	// the order of these substitutions is important, but it is hard to read. there must be a better way.

	// rand makes a distinction between â for /a/ before single consonants or vowels, ă for /a/ before clusters
	for charIndex, character := range OutputWords.Rand {
		if charIndex < len(OutputWords.Rand)-2 {
			if string(character) == "a" {
				if (IsConsonant(string(OutputWords.Rand[charIndex+1])) || IsSonorant(string(OutputWords.Rand[charIndex+1]))) && (IsConsonant(string(OutputWords.Rand[charIndex+2])) || IsSonorant(string(OutputWords.Rand[charIndex+2]))) {
					OutputWords.Rand = fmt.Sprintf("%s?%s", string([]rune(OutputWords.Rand[:charIndex])), string([]rune(OutputWords.Rand[charIndex+1:]))) // ? is a placeholder for ă, which is two bytes!
				} else {
					OutputWords.Rand = fmt.Sprintf("%s|%s", string([]rune(OutputWords.Rand[:charIndex])), string([]rune(OutputWords.Rand[charIndex+1:]))) // | is a placeholder for â, which is also two bytes
				}
			} else if string(character) == "i" {
				if (IsConsonant(string(OutputWords.Rand[charIndex+1])) || IsSonorant(string(OutputWords.Rand[charIndex+1]))) && (IsConsonant(string(OutputWords.Rand[charIndex+2])) || IsSonorant(string(OutputWords.Rand[charIndex+2]))) {
					OutputWords.Rand = fmt.Sprintf("%s(%s", string([]rune(OutputWords.Rand[:charIndex])), string([]rune(OutputWords.Rand[charIndex+1:]))) // ( is a placeholder for ĭ, which is two bytes!
				} else {
					OutputWords.Rand = fmt.Sprintf("%s)%s", string([]rune(OutputWords.Rand[:charIndex])), string([]rune(OutputWords.Rand[charIndex+1:]))) // ) is a placeholder for e so it does not interfere elsewhere
				}
			}
		}
	}

	OutputWords.Rand = strings.Replace(OutputWords.Rand, "cc", "c", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "qq", "q", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "q=", "=", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "ey", "ā", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "3y", "āā", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "*", "ŭ", -1)

	// replace temporary characters
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "?", "ă", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "|", "â", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "@", "a", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "e", "ĕ", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "3", "ā", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "o", "ŏ", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "u", "oo", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "@w", "oow", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "!w", "uu", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "iw", "u", -1)

	// replace temporary characters
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "(", "ĭ", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, ")", "e", -1)

	OutputWords.Rand = strings.Replace(OutputWords.Rand, "ăy", "ei", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "@y", "eei", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "!", "ee", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "ŏq", "ŏg", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "ăw", "ow", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "%", "ō", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "&", "oo", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "ăq", "ăg", -1) // rand uses k/g for /x/ after back vowels
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "aq", "ag", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "âq", "âg", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "ōq", "ōg", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "q", "h", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "c", "ch", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "6", "ŭl", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "7", "ŭn", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "+", "ŭm", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "8", "ŭm", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "9", "ŭn", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "0", "ŭl", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "#", "gw", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "$", "kw", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "=", "gw", -1)
	OutputWords.Rand = strings.Replace(OutputWords.Rand, "-", "", -1)

	// lexicon
	OutputWords.Lexicon = OutputWords.FrancisSmith
	OutputWords.Lexicon = strings.Replace(OutputWords.Lexicon, "'", ":", -1)
	OutputWords.Lexicon = strings.Replace(OutputWords.Lexicon, "l:", "l'", -1) // have to reconvert "l:" etc. probably easier to do this than to be smarter about the previous line
	OutputWords.Lexicon = strings.Replace(OutputWords.Lexicon, "n:", "n'", -1)
	OutputWords.Lexicon = strings.Replace(OutputWords.Lexicon, "m:", "m'", -1)

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
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "+", "êm", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "8", "êm", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "9", "ên", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "0", "êl", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "#", "gw", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "$", "kw", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "=", "qw", -1)
	OutputWords.Metallic = strings.Replace(OutputWords.Metallic, "-", "", -1)

	return OutputWords
}
