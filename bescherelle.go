// a program that conjugates mi'kmaw verbs and outputs them to an html template
// functions are (somewhat) thoroughly commented and so should be readable with reasonable previous knowledge of go and mi'kmaw grammar

// todo
// update localization

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type VerbType int

const (
	VII VerbType = iota // VII — verb inanimate intransitive
	VAI                 // VAI — verb animate intransitive
	VTI                 // VTI — verb transitive inanimate
	VTA                 // VTA — verb transitive animate
)

type Verb struct {
	Stem               string
	ContractedStem     string
	Conjugation        int
	ConjugationVariant string
	Type               VerbType
}

type Category struct { // a struct for reading the conjugation dictionary JSON
	Title string   `json:"title"`
	Forms []string `json:"forms"`
}

type Locale struct { // a struct for reading the conjugation dictionary JSON
	Language                    string   `json:"language"`
	TableTitles                 []string `json:"tabletitles"`
	SubjectPronouns             []string `json:"subjectpronouns"`
	InanimateObjectPronouns     []string `json:"inanobjpronouns"`
	SubjectObjectSplit          string   `json:"subjobjsplit"` // maybe not self explanatory — holds the "↓subject/object→" locale
	SubjectPronounsVTA          []string `json:"subjectpronounsvta"`
	ObjectPronounsVTA           []string `json:"objectpronounsvta"`
	PageTitle                   string   `json:"pagetitle"`
	OrthographyTooltip          string   `json:"orthographytooltip"`
	EntryPrompt                 string   `json:"entryprompt"`
	SummaryDetails              string   `json:"summarydetails"`
	LanguageFieldLabel          string   `json:"languagefieldlabel"`
	English                     string   `json:"english"`
	Mikmaw                      string   `json:"mikmaw"`
	French                      string   `json:"french"`
	ConjugateButton             string   `json:"conjugatebutton"`
	OutputConjugation           string   `json:"outputconjugation"`
	OutputModel                 string   `json:"outputmodel"`
	OutputVerbUnrecognized      string   `json:"outputverbunrecognized"`
	OutputTitle                 string   `json:"outputtitle"`
	ContactTitle                string   `json:"contacttitle"`
	ContactMe                   string   `json:"contactme"`
	InfoTitle                   string   `json:"infotitle"`
	HelpTitle                   string   `json:"helptitle"`
	HelpField                   string   `json:"helpfield"`
	SourceTitle                 string   `json:"sourcetitle"`
	SourceField                 string   `json:"sourcefield"`
	OrthographyRadioButtonTitle string   `json:"orthographyradiobuttontitle"`
	ElietDisclaimer             string   `json:"elietdisclaimer"`
	NestikDisclaimer            string   `json:"nestikdisclaimer"`
	NenkDisclaimer              string   `json:"nenkdisclaimer"`
	PewaqDisclaimer             string   `json:"pewaqdisclaimer"`
	PesatlDisclaimer            string   `json:"pesatldisclaimer"`
	KetukDisclaimer             string   `json:"ketukdisclaimer"`
	EykDisclaimer               string   `json:"eykdisclaimer"`
	VIIDisclaimer               string   `json:"viidisclaimer"`
	EwniaqDisclaimer            string   `json:"ewniaqdisclaimer"`
}

type Data struct { // for collecting the data of all tables
	Tables []Table
}

type Table struct { // for holding the one table
	Title          string
	Type           VerbType
	RowsAndColumns [][]string
}

type DisclaimerType struct { // this holds whether there is a disclaimer (Defined, bool), and what it is (DisclaimerText)
	Defined        bool
	DisclaimerText string
}

type MainPage struct { // this is what will be sent to the page
	Title                       string
	OrthographyTooltip          string
	EntryPrompt                 string
	SummaryDetails              string
	LanguageFieldLabel          string
	English                     string
	Mikmaw                      string
	French                      string
	ConjugateButton             string
	OutputConjugationTitle      string
	OutputConjugation           string
	OutputModelTitle            string
	OutputModel                 string
	OutputTitle                 string
	InputString                 string
	InfoTitle                   string
	HelpTitle                   string
	HelpField                   string
	SourceTitle                 string
	SourceField                 string
	ContactTitle                string
	ContactMe                   string
	OrthographyRadioButtonTitle string
	Disclaimer                  DisclaimerType
	TableData                   Data
}

var ConjugationDictionary = []Category{} // define a global conjugation dictionary to hold the readout of the .json file
var LocalizationDictionary = []Locale{}  // define a global localization lookup for all strings

func main() {
	conjugationDictionaryFile, errFileOpen := os.Open("conjdict.json") // open the json file
	if errFileOpen != nil {                                            // if there is an error
		fmt.Println(errFileOpen)
	}
	fmt.Println("Successfully opened conjdict.json.")

	defer conjugationDictionaryFile.Close() // defer closing until we are done using it

	conjugationDictionaryBytes, errFileRead := os.ReadFile("conjdict.json") // read the file into a byte array
	if errFileRead != nil {                                                 // if there is an error
		fmt.Println(errFileRead)
	}
	fmt.Println("Successfully read conjdict.json.")

	json.Unmarshal(conjugationDictionaryBytes, &ConjugationDictionary) // use the json package to unmarshal the byte array into the conjugation dictionary

	localizationFile, errFileOpen := os.Open("localization.json") // open the json file
	if errFileOpen != nil {                                       // if there is an error
		fmt.Println(errFileOpen)
	}
	fmt.Println("Successfully opened localization.json.")

	defer localizationFile.Close() // defer closing until we are done using it

	localizationBytes, errFileRead := os.ReadFile("localization.json") // read the file into a byte array
	if errFileRead != nil {                                            // if there is an error
		fmt.Println(errFileRead)
	}
	fmt.Println("Successfully read localization.json.")

	json.Unmarshal(localizationBytes, &LocalizationDictionary) // use the json package to unmarshal the byte array into the conjugation dictionary

	fileServe := http.FileServer(http.Dir("./assets"))              // add a stylesheet
	http.Handle("/assets/", http.StripPrefix("/assets", fileServe)) // no idea what this actually does, but this is from golang example code
	http.HandleFunc("/eng", engIndexHandler)                        // create a webpage for english
	http.HandleFunc("/mkw", mkwIndexHandler)                        // create a webpage for mi'kmaw
	http.HandleFunc("/fre", freIndexHandler)                        // create a webpage for mi'kmaw

	http.ListenAndServe(":8080", nil) // listen and serve
}

// this is the function that handles displaying the webpage for english
func engIndexHandler(writer http.ResponseWriter, reader *http.Request) {
	var WriteData Data                    // the tables to be sent to the template
	var page MainPage                     // all the fields that get passed to the template (incl. WriteData)
	var languageChoice string = "ENGL"    // string to get localization in english
	if reader.Method == http.MethodPost { // if the "submit/conjugate" button is pressed
		var InputStr string
		InputStr = reader.FormValue("verbinput")                        // get the input string
		var ConjugationArray [][]string                                 // load a conjugation array
		var InputVerb Verb                                              // load an InputVerb Verb type
		orthographyChoice := reader.FormValue("orthographyradiobutton") // a string value correstponding to the orthography chosen by the user
		// 0 = francis smith
		// 1 = listuguj
		if InputStr != "" { // if the input is not empty
			if orthographyChoice == "1" {
				InputStr = convertListugujtoFrancisSmith(InputStr) // if the user has chosen listuguj orthography, convert it to francis smith to run the program
			}
			ConjugationArray, InputVerb = readoutVerb(InputStr) // fill the conjugation array and input verb structs
			if orthographyChoice == "1" {
				ConjugationArray = convertFrancisSmithtoListuguj(ConjugationArray) // if the user has chosen listuguj orthography, convert all tables to listuguj
			}
			// make the tables differently for each verb type (VII, VAI, VTI, VTA) — point to a different function each time to do this
			if InputVerb.Type == VAI {
				WriteData = makeTablesVAI(ConjugationArray, languageChoice) // make the tables with this array based on localization language
			} else if InputVerb.Type == VTI {
				WriteData = makeTablesVTI(ConjugationArray, languageChoice) // make the tables with this array based on localization language
			} else if InputVerb.Type == VTA {
				WriteData = makeTablesVTA(ConjugationArray, languageChoice) // make the tables with this array based on localization language
			} else if InputVerb.Type == VII {
				WriteData = makeTablesVII(ConjugationArray, languageChoice) // make the tables with this array based on localization language
			}
		}
		page.OutputConjugation, page.OutputModel, page.Disclaimer = localizeOutput(languageChoice, InputVerb) // localize the output (get the conjugation, model, and disclaimers)
		page.InputString = InputStr                                                                           // the input string to be sent to the page (to be displayed as "you entered:")
	} else { // if the button was not pressed (i.e. on first load of the page without cache)
		ConjugationArray, InputVerb := readoutVerb("teluisit")                                                // load the conjugation array and InputVerb for "teluisit" as a default (Pacifique's first conjugation model)
		WriteData = makeTablesVAI(ConjugationArray, languageChoice)                                           // make the tables with this array based on localization language
		page.OutputConjugation, page.OutputModel, page.Disclaimer = localizeOutput(languageChoice, InputVerb) // localize the output (get the conjugation, model, and disclaimers)
		page.InputString = "teluisit"                                                                         // the input string is "teluisit"
	}
	page = localize(page, languageChoice) // localize everything else in the page (title, buttons, etc.)
	page.TableData = WriteData            // the tabledata is writedata (load the tables into the struct to be sent to the template)

	template, templateBuildErr := template.ParseFiles("template.html") // parse template.html
	if templateBuildErr != nil {                                       // if an error is thrown
		fmt.Println(templateBuildErr)
	}
	template.Execute(writer, page) // execute the template
}

// this is the same as engIndexHandler but with languageChoice MKMW (i.e. in mi'kmaw)
func mkwIndexHandler(writer http.ResponseWriter, reader *http.Request) {
	var WriteData Data                    // the tables to be sent to the template
	var page MainPage                     // all the fields that get passed to the template (incl. WriteData)
	var languageChoice string = "MKMW"    // string to get localization in english
	if reader.Method == http.MethodPost { // if the "submit/conjugate" button is pressed
		var InputStr string
		InputStr = reader.FormValue("verbinput")                        // get the input string
		var ConjugationArray [][]string                                 // load a conjugation array
		var InputVerb Verb                                              // load an InputVerb Verb type
		orthographyChoice := reader.FormValue("orthographyradiobutton") // a string value correstponding to the orthography chosen by the user
		// 0 = francis smith
		// 1 = listuguj
		if InputStr != "" { // if the input is not empty
			if orthographyChoice == "1" {
				InputStr = convertListugujtoFrancisSmith(InputStr) // if the user has chosen listuguj orthography, convert it to francis smith to run the program
			}
			ConjugationArray, InputVerb = readoutVerb(InputStr) // fill the conjugation array and input verb structs
			if orthographyChoice == "1" {
				ConjugationArray = convertFrancisSmithtoListuguj(ConjugationArray) // if the user has chosen listuguj orthography, convert all tables to listuguj
			}
			// make the tables differently for each verb type (VII, VAI, VTI, VTA) — point to a different function each time to do this
			if InputVerb.Type == VAI {
				WriteData = makeTablesVAI(ConjugationArray, languageChoice) // make the tables with this array based on localization language
			} else if InputVerb.Type == VTI {
				WriteData = makeTablesVTI(ConjugationArray, languageChoice) // make the tables with this array based on localization language
			} else if InputVerb.Type == VTA {
				WriteData = makeTablesVTA(ConjugationArray, languageChoice) // make the tables with this array based on localization language
			} else if InputVerb.Type == VII {
				WriteData = makeTablesVII(ConjugationArray, languageChoice) // make the tables with this array based on localization language
			}
		}
		page.OutputConjugation, page.OutputModel, page.Disclaimer = localizeOutput(languageChoice, InputVerb) // localize the output (get the conjugation, model, and disclaimers)
		page.InputString = InputStr                                                                           // the input string to be sent to the page (to be displayed as "you entered:")
	} else { // if the button was not pressed (i.e. on first load of the page without cache)
		ConjugationArray, InputVerb := readoutVerb("teluisit")                                                // load the conjugation array and InputVerb for "teluisit" as a default (Pacifique's first conjugation model)
		WriteData = makeTablesVAI(ConjugationArray, languageChoice)                                           // make the tables with this array based on localization language
		page.OutputConjugation, page.OutputModel, page.Disclaimer = localizeOutput(languageChoice, InputVerb) // localize the output (get the conjugation, model, and disclaimers)
		page.InputString = "teluisit"                                                                         // the input string is "teluisit"
	}
	page = localize(page, languageChoice) // localize everything else in the page (title, buttons, etc.)
	page.TableData = WriteData            // the tabledata is writedata (load the tables into the struct to be sent to the template)

	template, templateBuildErr := template.ParseFiles("template.html") // parse template.html
	if templateBuildErr != nil {                                       // if an error is thrown
		fmt.Println(templateBuildErr)
	}
	template.Execute(writer, page) // execute the template
}

// this is the same as engIndexHandler but with languageChoice FREN (i.e. in french)
func freIndexHandler(writer http.ResponseWriter, reader *http.Request) {
	var WriteData Data                    // the tables to be sent to the template
	var page MainPage                     // all the fields that get passed to the template (incl. WriteData)
	var languageChoice string = "FREN"    // string to get localization in english
	if reader.Method == http.MethodPost { // if the "submit/conjugate" button is pressed
		var InputStr string
		InputStr = reader.FormValue("verbinput")                        // get the input string
		var ConjugationArray [][]string                                 // load a conjugation array
		var InputVerb Verb                                              // load an InputVerb Verb type
		orthographyChoice := reader.FormValue("orthographyradiobutton") // a string value correstponding to the orthography chosen by the user
		// 0 = francis smith
		// 1 = listuguj
		if InputStr != "" { // if the input is not empty
			if orthographyChoice == "1" {
				InputStr = convertListugujtoFrancisSmith(InputStr) // if the user has chosen listuguj orthography, convert it to francis smith to run the program
			}
			ConjugationArray, InputVerb = readoutVerb(InputStr) // fill the conjugation array and input verb structs
			if orthographyChoice == "1" {
				ConjugationArray = convertFrancisSmithtoListuguj(ConjugationArray) // if the user has chosen listuguj orthography, convert all tables to listuguj
			}
			// make the tables differently for each verb type (VII, VAI, VTI, VTA) — point to a different function each time to do this
			if InputVerb.Type == VAI {
				WriteData = makeTablesVAI(ConjugationArray, languageChoice) // make the tables with this array based on localization language
			} else if InputVerb.Type == VTI {
				WriteData = makeTablesVTI(ConjugationArray, languageChoice) // make the tables with this array based on localization language
			} else if InputVerb.Type == VTA {
				WriteData = makeTablesVTA(ConjugationArray, languageChoice) // make the tables with this array based on localization language
			} else if InputVerb.Type == VII {
				WriteData = makeTablesVII(ConjugationArray, languageChoice) // make the tables with this array based on localization language
			}
		}
		page.OutputConjugation, page.OutputModel, page.Disclaimer = localizeOutput(languageChoice, InputVerb) // localize the output (get the conjugation, model, and disclaimers)
		page.InputString = InputStr                                                                           // the input string to be sent to the page (to be displayed as "you entered:")
	} else { // if the button was not pressed (i.e. on first load of the page without cache)
		ConjugationArray, InputVerb := readoutVerb("teluisit")                                                // load the conjugation array and InputVerb for "teluisit" as a default (Pacifique's first conjugation model)
		WriteData = makeTablesVAI(ConjugationArray, languageChoice)                                           // make the tables with this array based on localization language
		page.OutputConjugation, page.OutputModel, page.Disclaimer = localizeOutput(languageChoice, InputVerb) // localize the output (get the conjugation, model, and disclaimers)
		page.InputString = "teluisit"                                                                         // the input string is "teluisit"
	}
	page = localize(page, languageChoice) // localize everything else in the page (title, buttons, etc.)
	page.TableData = WriteData            // the tabledata is writedata (load the tables into the struct to be sent to the template)

	template, templateBuildErr := template.ParseFiles("template.html") // parse template.html
	if templateBuildErr != nil {                                       // if an error is thrown
		fmt.Println(templateBuildErr)
	}
	template.Execute(writer, page) // execute the template
}

func IsConsonant(category string) bool { // returns true if the passed slice is in this list
	switch category {
	case
		"j",
		"k",
		"l",
		"m",
		"n",
		"p",
		"q",
		"s",
		"t",
		"w",
		"y":
		return true
	}
	return false
}

func IsPlosive(category string) bool { // returns true if the passed slice is in this list
	switch category {
	case
		"j",
		"k",
		"p",
		"q",
		"s",
		"t":
		return true
	}
	return false
}

// converts anything written in listuguj orthography into francis-smith
func convertListugujtoFrancisSmith(InputStr string) string {
	OutputStr := strings.ToLower(InputStr)
	// everything below is just replacing character combinations with others
	// the output string is in francis-smith
	if strings.Contains(OutputStr, "g") == true {
		OutputStr = strings.Replace(OutputStr, "g", "k", -1)
	}
	if strings.Contains(OutputStr, "ai") == true {
		OutputStr = strings.Replace(OutputStr, "ai", "ay", -1)
	}
	if strings.Contains(OutputStr, "a'i") == true {
		OutputStr = strings.Replace(OutputStr, "a'i", "a'y", -1)
	}
	if strings.Contains(OutputStr, "ei") == true {
		OutputStr = strings.Replace(OutputStr, "ei", "ey", -1)
	}
	if strings.Contains(OutputStr, "e'i") == true {
		OutputStr = strings.Replace(OutputStr, "e'i", "e'y", -1)
	}
	// replace all apostrophes after consonants as schwas, but use the escape character "*"
	if strings.Contains(OutputStr, "j'") == true {
		OutputStr = strings.Replace(OutputStr, "j'", "j*", -1)
	}
	if strings.Contains(OutputStr, "k'") == true {
		OutputStr = strings.Replace(OutputStr, "k'", "k*", -1)
	}
	if strings.Contains(OutputStr, "m'") == true {
		OutputStr = strings.Replace(OutputStr, "m'", "m*", -1)
	}
	if strings.Contains(OutputStr, "n'") == true {
		OutputStr = strings.Replace(OutputStr, "n'", "n*", -1)
	}
	if strings.Contains(OutputStr, "p'") == true {
		OutputStr = strings.Replace(OutputStr, "p'", "p*", -1)
	}
	if strings.Contains(OutputStr, "q'") == true {
		OutputStr = strings.Replace(OutputStr, "q'", "q*", -1)
	}
	if strings.Contains(OutputStr, "s'") == true {
		OutputStr = strings.Replace(OutputStr, "s'", "s*", -1)
	}
	if strings.Contains(OutputStr, "t'") == true {
		OutputStr = strings.Replace(OutputStr, "t'", "t*", -1)
	}
	return OutputStr
}

// converts anything written in francis-smith into listuguj
func convertFrancisSmithtoListuguj(InputArray [][]string) [][]string {
	for sliceIndex := range InputArray {
		for stringIndex, string := range InputArray[sliceIndex] {
			outputStr := string
			if strings.Contains(outputStr, "k") == true {
				outputStr = strings.Replace(outputStr, "k", "g", -1)
			}
			if strings.Contains(outputStr, "ɨ") == true {
				outputStr = strings.Replace(outputStr, "ɨ", "'", -1)
			}
			if strings.Contains(outputStr, "y") == true {
				outputStr = strings.Replace(outputStr, "y", "i", -1)
			}
			// converting "y" to "i" in strings of "yi" will lead to "ii"
			if strings.Contains(outputStr, "ii") == true {
				outputStr = strings.Replace(outputStr, "ii", "i", -1)
			}
			InputArray[sliceIndex][stringIndex] = outputStr
		}
	}
	return InputArray
}

// this function returns the conjugation number and a model for the verb based off the InputVerb struct
// it will also tell the template to put a disclaimer for those verbs that should have it
// pretty self-explanatory, so not much commenting here
func localizeOutput(languageChoice string, InputVerb Verb) (string, string, DisclaimerType) {
	var LocalOutputConjugation string
	var LocalOutputModel string
	var LocalDisclaimer DisclaimerType
	LocalDisclaimer.Defined = false                   // this tells the template whether or not to display a tooltip with a disclaimer for a particular verb group
	for _, language := range LocalizationDictionary { // loop through the objects in the localization file (ENGL or MKMW)
		if language.Language == languageChoice {
			if InputVerb.Conjugation == 1 { // cannot use this number directly; there are some that are "between" conjugations
				if InputVerb.ConjugationVariant == "asit" {
					LocalOutputConjugation = "1"
					LocalOutputModel = "pejila'sit"
				} else if InputVerb.ConjugationVariant == "asik" {
					LocalOutputConjugation = "1"
					LocalOutputModel = "enqa'sik"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.VIIDisclaimer // for multiple forms in the future (VII only)
				} else if InputVerb.ConjugationVariant == "ink" {
					LocalOutputConjugation = "1"
					LocalOutputModel = "pekisink"
				} else if InputVerb.ConjugationVariant == "inan" {
					LocalOutputConjugation = "1"
					LocalOutputModel = "maqatkwik"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.VIIDisclaimer // for multiple forms in the future (VII only)
				} else if InputVerb.ConjugationVariant == "std" {
					LocalOutputConjugation = "1"
					LocalOutputModel = "teluisit"
				}
			} else if InputVerb.Conjugation == 2 {
				if InputVerb.ConjugationVariant == "long" {
					LocalOutputConjugation = "2"
					LocalOutputModel = "ajipuna't"
				} else if InputVerb.ConjugationVariant == "inan" {
					LocalOutputConjugation = "2"
					LocalOutputModel = "pesaq"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.VIIDisclaimer // for multiple forms in the future (VII only)
				} else if InputVerb.ConjugationVariant == "diph" {
					LocalOutputConjugation = "1~2"
					LocalOutputModel = "wekayk"
				} else if InputVerb.ConjugationVariant == "std" {
					LocalOutputConjugation = "2"
					LocalOutputModel = "amalkat"
				}
			} else if InputVerb.Conjugation == 3 {
				if InputVerb.ConjugationVariant == "iet" {
					LocalOutputConjugation = "3"
					LocalOutputModel = "eliet"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.ElietDisclaimer // variant forms for land/water travel in the dual
				} else if InputVerb.ConjugationVariant == "iaq" {
					LocalOutputConjugation = "3"
					LocalOutputModel = "ewniaq"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.EwniaqDisclaimer // variant forms for land/water travel in the dual (inanimate)
				} else if InputVerb.ConjugationVariant == "uet" {
					LocalOutputConjugation = "3"
					LocalOutputModel = "teluet"
				} else if InputVerb.ConjugationVariant == "eket" {
					LocalOutputConjugation = "3"
					LocalOutputModel = "teweket"
				} else if InputVerb.ConjugationVariant == "long" {
					LocalOutputConjugation = "1~3"
					LocalOutputModel = "wele'k"
				} else if InputVerb.ConjugationVariant == "inan" {
					LocalOutputConjugation = "3"
					LocalOutputModel = "te'sipunqek"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.VIIDisclaimer // for multiple forms in the future (VII only)
				} else if InputVerb.ConjugationVariant == "std" {
					LocalOutputConjugation = "3"
					LocalOutputModel = "ewi'kiket"
				}
			} else if InputVerb.Conjugation == 4 {
				if InputVerb.ConjugationVariant == "ibar" {
					LocalOutputConjugation = "4"
					LocalOutputModel = "nestɨk"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.NestikDisclaimer // about the variant -kik/-mi'tij forms in the 3rd person plural
				} else if InputVerb.ConjugationVariant == "estem" {
					LocalOutputConjugation = "4"
					LocalOutputModel = "telte'k"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.NestikDisclaimer // about the variant -kik/-mi'tij forms in the 3rd person plural
				} else if InputVerb.ConjugationVariant == "cons" {
					LocalOutputConjugation = "4"
					LocalOutputModel = "nenk"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.NenkDisclaimer // some verbs in this group are inanimate subject only
				} else if InputVerb.ConjugationVariant == "istem" {
					LocalOutputConjugation = "4"
					LocalOutputModel = "ketkwi'k"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.NestikDisclaimer // about the variant -kik/-mi'tij forms in the 3rd person plural
				} else if InputVerb.ConjugationVariant == "eyk" {
					LocalOutputConjugation = "4"
					LocalOutputModel = "eyk"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.EykDisclaimer // explaining that eyk is animate, etek is inanimate
				} else if InputVerb.ConjugationVariant == "astem" {
					LocalOutputConjugation = "4"
					LocalOutputModel = "pewa'q"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.PewaqDisclaimer // about the variant -kik/-mi'tij forms in the 3rd person plural
				} else if InputVerb.ConjugationVariant == "kstem" {
					LocalOutputConjugation = "4"
					LocalOutputModel = "ewi'kɨk"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.NestikDisclaimer // about the variant -kik/-mi'tij forms in the 3rd person plural
				} else if InputVerb.ConjugationVariant == "inan" {
					LocalOutputConjugation = "4~5"
					LocalOutputModel = "telamu'k"
				} else if InputVerb.ConjugationVariant == "std" {
					LocalOutputConjugation = "4"
					LocalOutputModel = "kesatk"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.NestikDisclaimer // about the variant -kik/-mi'tij forms in the 3rd person plural
				}
			} else if InputVerb.Conjugation == 5 {
				if InputVerb.ConjugationVariant == "kuk" {
					LocalOutputConjugation = "5"
					LocalOutputModel = "ketuk"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.KetukDisclaimer // some verbs in this group are inanimate subject only
				} else if InputVerb.ConjugationVariant == "std" {
					LocalOutputConjugation = "5"
					LocalOutputModel = "mena'toq"
				}
			} else if InputVerb.Conjugation == 6 {
				if InputVerb.ConjugationVariant == "istem" {
					LocalOutputConjugation = "6"
					LocalOutputModel = "nemiatl"
				} else if InputVerb.ConjugationVariant == "aestem" {
					LocalOutputConjugation = "6"
					LocalOutputModel = "pesa'tl"
					LocalDisclaimer.Defined = true
					LocalDisclaimer.DisclaimerText = language.PesatlDisclaimer // some verbs of this group have -a- stems, some have -e- stems
				} else if InputVerb.ConjugationVariant == "ibar" {
					LocalOutputConjugation = "6"
					LocalOutputModel = "e'natl"
				} else if InputVerb.ConjugationVariant == "std" {
					LocalOutputConjugation = "6"
					LocalOutputModel = "kesalatl"
				}
			} else if InputVerb.Conjugation == 7 {
				LocalOutputConjugation = "7"
				LocalOutputModel = "kisituatl"
			} else {
				LocalOutputConjugation = ""
				LocalOutputModel = language.OutputVerbUnrecognized
			}
		}
	}
	return LocalOutputConjugation, LocalOutputModel, LocalDisclaimer
}

// this function returns the proper strings for titles, buttons, tenses, subject persons, etc. based on language
func localize(page MainPage, languageChoice string) MainPage {
	// get localization strings
	for _, language := range LocalizationDictionary { // loop through the objects in the localization file (ENGL or MKMW)
		if language.Language == languageChoice {
			page.Title = language.PageTitle
			page.OrthographyTooltip = language.OrthographyTooltip
			page.EntryPrompt = language.EntryPrompt
			page.SummaryDetails = language.SummaryDetails
			page.LanguageFieldLabel = language.LanguageFieldLabel
			page.English = language.English
			page.Mikmaw = language.Mikmaw
			page.French = language.French
			page.ConjugateButton = language.ConjugateButton
			page.OutputConjugationTitle = language.OutputConjugation
			page.OutputModelTitle = language.OutputModel
			page.OutputTitle = language.OutputTitle
			page.InfoTitle = language.InfoTitle
			page.HelpTitle = language.HelpTitle
			page.HelpField = language.HelpField
			page.SourceTitle = language.SourceTitle
			page.SourceField = language.SourceField
			page.OrthographyRadioButtonTitle = language.OrthographyRadioButtonTitle
			page.ContactTitle = language.ContactTitle
			page.ContactMe = language.ContactMe
		}
	}
	return page
}

// this function passes to other functions to return an InputVerb of type Verb and a ConjugationArray of a multidimensional string slice
func readoutVerb(InputStr string) ([][]string, Verb) {
	var ConjugationArray [][]string       // a composite literal of strings to hold the conjugated forms
	InputVerb, err := parseVerb(InputStr) // parse the verb stem
	if err != nil {                       // if the parseVerb function throws an error
		fmt.Println(err)
		return ConjugationArray, InputVerb
	}
	InputVerb.ContractedStem = contractStem(InputVerb.Stem, InputVerb.Conjugation) // get the contracted stem

	ConjugationArray = conjugateVerb(InputVerb) // initialize the composite literal multidimensional string slice

	return ConjugationArray, InputVerb
}

// the rows and columns need to be switched:
// the backend runs on columns — it is much easier to do manipulation by column than by row
// html tables work by rows, so they have to be switched
// (this can be done with css on the frontend, but it runs into problems with tables that are too long)
func transposeRowsAndColumns(InputArray [][]string) [][]string {
	arrayLength := len(InputArray[0])                             // get the length of a row as the length of the first slice of the input array
	temporaryArray := make([][]string, arrayLength)               // make a temporary array of the same length as the input array
	for sliceIndex := 0; sliceIndex < arrayLength; sliceIndex++ { // for each slice
		sliceLength := len(InputArray)                                   // get the length of each slice as the length of the input array
		temporaryArray[sliceIndex] = make([]string, sliceLength)         // make a slice in the temporary array of the same length as the input array slice
		for stringIndex := 0; stringIndex < sliceLength; stringIndex++ { // for each slice
			temporaryArray[sliceIndex][stringIndex] = InputArray[stringIndex][sliceIndex] // populate the temporary array with the input array values
		}
	}
	return temporaryArray
}

// this makes the tables for VAI verbs
func makeTablesVAI(ConjugationArray [][]string, languageChoice string) Data {
	// get localization strings
	var subjectPronounsVAI []string                   // subject pronouns
	var tensesVAI []string                            // tense headers (e.g. present, present negative)
	for _, language := range LocalizationDictionary { // loop through the objects in the localization file (ENGL or MKMW)
		if language.Language == languageChoice {
			subjectPronounsVAI = language.SubjectPronouns
			tensesVAI = language.TableTitles
		}
	}

	var OutputData Data
	// for each slice in the array, make a table from it
	for sliceIndex, slice := range ConjugationArray {
		var CurrentTable Table
		CurrentTable.Type = VAI
		// need to filter the subject pronouns
		var subjectPronounsNarrowed []string
		if sliceIndex < 2 { // the present tenses
			subjectPronounsNarrowed = subjectPronounsVAI
		} else if (sliceIndex >= 2 && sliceIndex <= 7) || (sliceIndex >= 16 && sliceIndex <= 21) { // the past and if conjunct tenses
			for index, item := range subjectPronounsVAI {
				if index != 4 && index != 5 && index != 6 &&
					index != 13 && index != 14 && index != 15 &&
					index != 21 && index != 22 && index != 23 { // remove unused persons
					subjectPronounsNarrowed = append(subjectPronounsNarrowed, item)
				}
			}
		} else if (sliceIndex >= 8 && sliceIndex <= 9) || (sliceIndex >= 22 && sliceIndex <= 25) { // future and conditional tenses
			for index, item := range subjectPronounsVAI {
				if index != 5 && index != 6 &&
					index != 14 && index != 15 &&
					index != 22 && index != 23 { // remove unused persons
					subjectPronounsNarrowed = append(subjectPronounsNarrowed, item)
				}
			}
		} else if sliceIndex >= 10 && sliceIndex <= 11 { // imperative tenses
			for index, item := range subjectPronounsVAI {
				if index != 0 && index != 4 && index != 5 && index != 6 &&
					index != 9 && index != 13 && index != 14 && index != 15 &&
					index != 17 && index != 21 && index != 22 && index != 23 { // remove unused persons
					subjectPronounsNarrowed = append(subjectPronounsNarrowed, item)
				}
			}
		} else if sliceIndex >= 12 && sliceIndex <= 15 { // when conjunct tenses
			for index, item := range subjectPronounsVAI {
				if index != 6 && index != 15 && index != 23 { // remove unused persons
					subjectPronounsNarrowed = append(subjectPronounsNarrowed, item)
				}
			}
		}
		CurrentTable.Title = tensesVAI[sliceIndex]                                                 // the table title is from localization.json
		CurrentTable.RowsAndColumns = append(CurrentTable.RowsAndColumns, subjectPronounsNarrowed) // append the subject pronouns as the first column
		CurrentTable.RowsAndColumns = append(CurrentTable.RowsAndColumns, slice)                   // append the current slice of the conjugation dictionary (corresponds to tense) as the second column
		CurrentTable.RowsAndColumns = transposeRowsAndColumns(CurrentTable.RowsAndColumns)         // switch the rows and columns for html template
		OutputData.Tables = append(OutputData.Tables, CurrentTable)                                // append the current table to the output data table slice
		subjectPronounsNarrowed = nil                                                              // filtered subject pronouns have to be reset for the next iteration
	}
	return OutputData
}

// this makes the tables for VTI verbs
func makeTablesVTI(ConjugationArray [][]string, languageChoice string) Data {
	// get localization strings
	var subjectPronounsVTI []string                   // subject pronouns
	var tensesVTI []string                            // tense headers (e.g. present, present negative)
	var objectPronounsVTI []string                    // object headers for inanimates (conj. 4, 5)
	var subjectObjectHeader []string                  // the (subject/object) header
	for _, language := range LocalizationDictionary { // loop through the objects in the localization file (ENGL or MKMW)
		if language.Language == languageChoice {
			subjectPronounsVTI = language.SubjectPronouns
			tensesVTI = language.TableTitles
			objectPronounsVTI = language.InanimateObjectPronouns
			subjectObjectHeader = append(subjectObjectHeader, language.SubjectObjectSplit) // the header goes first (↓subject/object→)
		}
	}

	var OutputData Data
	// for the present and past, make a table with singular/plural objects
	for formIndex := 0; formIndex < 8; formIndex++ {
		var CurrentTable Table
		CurrentTable.Type = VTI

		var splitForms2 [][]string
		joinedForms := strings.Join(ConjugationArray[formIndex], "%") // join all forms in the input slice with "%"
		splitForms1 := strings.Split(joinedForms, "%||%")             // split all forms in the input at "%&&%". this returns a string for each person
		for _, form := range splitForms1 {                            // for every form in the split list
			formSlice := strings.Split(form, "%")        // split them again at "%"
			splitForms2 = append(splitForms2, formSlice) // append the slice for each person to splitforms2
		}

		// need to filter the subject pronouns
		var subjectPronounsNarrowed []string
		if formIndex < 2 { // the present tenses
			subjectPronounsNarrowed = subjectPronounsVTI
		} else if formIndex >= 2 && formIndex <= 7 { // the past tenses
			for index, item := range subjectPronounsVTI {
				if index != 4 && index != 5 && index != 6 &&
					index != 13 && index != 14 && index != 15 &&
					index != 21 && index != 22 && index != 23 { // remove unused persons
					subjectPronounsNarrowed = append(subjectPronounsNarrowed, item)
				}
			}
		}

		subjectPronounsNarrowed = append(subjectObjectHeader, subjectPronounsNarrowed...) // add the subject/object header to the subject pronouns

		CurrentTable.Title = tensesVTI[formIndex]                                                  // the table title is the current element in the slice
		CurrentTable.RowsAndColumns = append(CurrentTable.RowsAndColumns, subjectPronounsNarrowed) // append the subject pronouns to the current table as the first column
		for columnIndex := range splitForms2 {                                                     // for each column in the split forms (one for singular objects, one for plural objects)
			var newColumn []string                                                       // create the current column as a string slice
			newColumn = append(newColumn, objectPronounsVTI[columnIndex])                // append the object pronouns as the first element of the column
			newColumn = append(newColumn, splitForms2[columnIndex]...)                   // append everything in the current splitforms slice (singular or plural)
			CurrentTable.RowsAndColumns = append(CurrentTable.RowsAndColumns, newColumn) // append the current column to the current table
		}
		CurrentTable.RowsAndColumns = transposeRowsAndColumns(CurrentTable.RowsAndColumns) // switch the rows and columns for the html template
		OutputData.Tables = append(OutputData.Tables, CurrentTable)                        // append the current table to the output data table slice
		subjectPronounsNarrowed = nil                                                      // the filtered subject pronouns have to be cleared for the next iteration
	}

	// tenses after the present and past have no objects
	for formIndex := 8; formIndex < len(ConjugationArray); formIndex++ {
		// need to filter the subject pronouns
		var CurrentTable Table
		CurrentTable.Type = VAI // these tables act like VAI tables
		var subjectPronounsNarrowed []string
		if (formIndex >= 8 && formIndex <= 9) || (formIndex >= 22 && formIndex <= 25) { // future and conditional tenses
			for index, item := range subjectPronounsVTI {
				if index != 5 && index != 6 &&
					index != 14 && index != 15 &&
					index != 22 && index != 23 { // remove unused persons
					subjectPronounsNarrowed = append(subjectPronounsNarrowed, item)
				}
			}
		} else if formIndex >= 10 && formIndex <= 11 { // imperative tenses
			for index, item := range subjectPronounsVTI {
				if index != 0 && index != 4 && index != 5 && index != 6 &&
					index != 9 && index != 13 && index != 14 && index != 15 &&
					index != 17 && index != 21 && index != 22 && index != 23 { // remove unused persons
					subjectPronounsNarrowed = append(subjectPronounsNarrowed, item)
				}
			}
		} else if formIndex >= 12 && formIndex <= 15 { // when conjunct tenses
			for index, item := range subjectPronounsVTI {
				if index != 6 && index != 15 && index != 23 { // remove unused persons
					subjectPronounsNarrowed = append(subjectPronounsNarrowed, item)
				}
			}
		} else if formIndex >= 16 && formIndex <= 21 { // if conjunct tenses
			for index, item := range subjectPronounsVTI {
				if index != 4 && index != 5 && index != 6 &&
					index != 13 && index != 14 && index != 15 &&
					index != 21 && index != 22 && index != 23 { // remove unused persons
					subjectPronounsNarrowed = append(subjectPronounsNarrowed, item)
				}
			}
		}
		CurrentTable.Title = tensesVTI[formIndex]                                                      // the table title is from localization.json
		CurrentTable.RowsAndColumns = append(CurrentTable.RowsAndColumns, subjectPronounsNarrowed)     // append the subject pronouns to the current table as the first column
		CurrentTable.RowsAndColumns = append(CurrentTable.RowsAndColumns, ConjugationArray[formIndex]) // append the current slice of the conjugation array as the second column
		CurrentTable.RowsAndColumns = transposeRowsAndColumns(CurrentTable.RowsAndColumns)             // switch the rows and columns for the html template
		OutputData.Tables = append(OutputData.Tables, CurrentTable)                                    // append the current table to the output data table slice
		subjectPronounsNarrowed = nil                                                                  // filtered subject pronouns have to be reset for the next iteration
	}
	return OutputData
}

// this makes the tables for VTA verbs
func makeTablesVTA(ConjugationArray [][]string, languageChoice string) Data {
	var OutputData Data

	// get localization strings
	var subjectPronounsVTA []string                   // subject pronouns
	var tensesVTA []string                            // tense headers (e.g. present, present negative)
	var objectPronounsVTA []string                    // object headers for inanimates (conj. 4, 5)
	var subjectObjectHeader []string                  // the (subject/object) header
	for _, language := range LocalizationDictionary { // loop through the objects in the localization file (ENGL or MKMW)
		if language.Language == languageChoice {
			subjectPronounsVTA = language.SubjectPronounsVTA
			tensesVTA = language.TableTitles
			objectPronounsVTA = language.ObjectPronounsVTA
			subjectObjectHeader = append(subjectObjectHeader, language.SubjectObjectSplit) // the header goes first (↓subject/object→)
		}
	}
	var objectPronounsVTANarrowed []string // make a string slice for storing filtered object pronouns
	for itemIndex, item := range objectPronounsVTA {
		if itemIndex != 3 && itemIndex != 8 { // filter items 3 and 8 out (absentative object)
			objectPronounsVTANarrowed = append(objectPronounsVTANarrowed, item)
		}
	}
	var SubjectPronounsVTANarrowed []string // make a string slice for storing filtered subject pronouns
	for itemIndex, item := range subjectPronounsVTA {
		if itemIndex != 0 && itemIndex != 4 && itemIndex != 7 && itemIndex != 11 { // filter items 0, 4, 7, 11 out (first person subject)
			SubjectPronounsVTANarrowed = append(SubjectPronounsVTANarrowed, item)
		}
	}
	for tableIndex := range ConjugationArray {
		var splitForms2 [][]string
		joinedForms := strings.Join(ConjugationArray[tableIndex], "%") // join all forms in the input slice with "%"
		splitForms1 := strings.Split(joinedForms, "%&&%")              // split all forms in the input at "%&&%". this returns a string for each person
		for _, form := range splitForms1 {                             // for every form in the split list
			formSlice := strings.Split(form, "%")        // split them again at "%"
			splitForms2 = append(splitForms2, formSlice) // append the slice for each person to splitforms2
		}
		var splitForms2Narrowed [][]string // a multidimensional slice to store a filtered version of splitforms2 — some tenses do not use the full slate of persons
		for itemIndex, item := range splitForms2 {
			if itemIndex != 3 && itemIndex != 8 { // items with index 3 and 8 should be filtered out (the absentative objects)
				splitForms2Narrowed = append(splitForms2Narrowed, item)
			}
		}

		var CurrentTable Table            // the current table
		var objectPronounsLocal []string  // a local object pronoun slice
		var subjectPronounsLocal []string // a local subject pronoun slice
		var splitFormsLocal [][]string    // a local split forms slice — some tenses (the imperative) should have some person references removed
		if tableIndex >= 8 {              // if not the present and past
			objectPronounsLocal = objectPronounsVTANarrowed // the local object pronouns are the filtered ones
		} else {
			objectPronounsLocal = objectPronounsVTA // the local object pronouns are the unfiltered ones
		}
		if tableIndex == 9 { // if it is the future negative
			splitFormsLocal = splitForms2Narrowed // local splitforms (rows/columns) is filtered — this removes the forms used in the present negative that aren't in the future negative
		} else {
			splitFormsLocal = splitForms2 // local rows/columns are unfiltered (come directly from conjdict.json)
		}
		if tableIndex == 10 || tableIndex == 11 { // if it is the imperative, imperative negative
			subjectPronounsLocal = SubjectPronounsVTANarrowed // the local subject pronouns are filtered (do not need first persons for this)
		} else {
			subjectPronounsLocal = subjectPronounsVTA // the local subject pronouns are unfiltered
		}

		subjectPronounsLocal = append(subjectObjectHeader, subjectPronounsLocal...) // append the (↓subject/object→) to the beginning of the subject pronouns

		if tableIndex >= 23 { // if it is past the 23rd table index (attestive conditional) — the attestive conditional for VTA verbs is NULL, but the string still exists in localization.json, so it needs to be skipped
			CurrentTable.Title = tensesVTA[tableIndex+1]
		} else {
			CurrentTable.Title = tensesVTA[tableIndex]
		}
		CurrentTable.Type = VTA                                                                 // set the table type so that the template parser knows what kind of table to display
		CurrentTable.RowsAndColumns = append(CurrentTable.RowsAndColumns, subjectPronounsLocal) // append the subject pronouns to the current table
		for columnIndex := range splitFormsLocal {                                              // for all columns in the local split forms (split at && in the string in conjdict.json)
			var newColumn []string                                                       // make a column
			newColumn = append(newColumn, objectPronounsLocal[columnIndex])              // add the object pronoun first
			newColumn = append(newColumn, splitFormsLocal[columnIndex]...)               // add everything in the column of local split forms
			CurrentTable.RowsAndColumns = append(CurrentTable.RowsAndColumns, newColumn) // append the whole column to the current table
		}
		CurrentTable.RowsAndColumns = transposeRowsAndColumns(CurrentTable.RowsAndColumns) // switch the rows and columns for the html template
		OutputData.Tables = append(OutputData.Tables, CurrentTable)                        // append the current table to the outputdata slice of tables
	}
	return OutputData
}

// this makes the tables for VII verbs
func makeTablesVII(ConjugationArray [][]string, languageChoice string) Data {
	// get localization strings
	var subjectPronounsVII []string                   // subject pronouns
	var subjectPronounsVIINarrowed []string           // subject pronouns without absentatives
	var subjectPronounsVIILocal []string              // subjectpronouns that are mobile and build the table
	var tensesVII []string                            // tense headers (e.g. present, present negative)
	var tempslice []string                            // a temporary string to store all of the subject pronouns, which will be filtered
	for _, language := range LocalizationDictionary { // loop through the objects in the localization file (ENGL or MKMW)
		if language.Language == languageChoice {
			tempslice = language.SubjectPronouns // this is the same subject pronouns as for VAI/VTI tables
			tensesVII = language.TableTitles
		}
	}

	for itemIndex, item := range tempslice { // filter only the inanimate subjects out of the total subject pronoun slice (with absentatives)
		if itemIndex == 3 || itemIndex == 6 || itemIndex == 12 || itemIndex == 15 || itemIndex == 20 || itemIndex == 23 { // items 3, 6, 12, 15, 20, and 23 are the inanimate subjects
			subjectPronounsVII = append(subjectPronounsVII, item)
		}
		if itemIndex == 3 || itemIndex == 12 || itemIndex == 20 { // items 3, 12, and 20 are the inanimate subjects
			subjectPronounsVIINarrowed = append(subjectPronounsVIINarrowed, item)
		}
	}

	var OutputData Data
	// for every slice in the conjugation array, make a table
	for sliceIndex, slice := range ConjugationArray {
		// in the present, present negative (i.e. slice < 2), use absentative pronouns, otherwise omit them
		if sliceIndex < 2 {
			subjectPronounsVIILocal = subjectPronounsVII
		} else {
			subjectPronounsVIILocal = subjectPronounsVIINarrowed
		}

		var CurrentTable Table
		CurrentTable.Type = VII
		if sliceIndex >= 23 { // skip the 23rd (attestive conditional, this form does not exist for inanimates)
			CurrentTable.Title = tensesVII[sliceIndex+1]
		} else {
			CurrentTable.Title = tensesVII[sliceIndex]
		}
		CurrentTable.RowsAndColumns = append(CurrentTable.RowsAndColumns, subjectPronounsVIILocal) // append the subject pronouns as the first row
		CurrentTable.RowsAndColumns = append(CurrentTable.RowsAndColumns, slice)                   // append the current slice in the conjugation array (sliceIndex corresponds to the tense)
		CurrentTable.RowsAndColumns = transposeRowsAndColumns(CurrentTable.RowsAndColumns)         // switch rows and columns for the html template
		OutputData.Tables = append(OutputData.Tables, CurrentTable)                                // append each table to the output data table slice
	}
	return OutputData
}

// this will return a two-dimensional string slice (each string is a verb form, each slice of string is a tense, the whole thing is a slice of tenses)
// is called in the readoutVerb function
func conjugateVerb(InputVerb Verb) [][]string {
	var Namespace string        // the string in the index that corresponds to the variant object
	var FormIndex string        // the whole index that points to the correct object
	var OutputArray [][]string  // the array of forms that are gathered
	var readErr error           // if the reader throws an error
	var temporaryForms []string // for doing manipulation of forms

	// point to the correct variant in the file
	if (InputVerb.ConjugationVariant == "estem" || InputVerb.ConjugationVariant == "cons" ||
		InputVerb.ConjugationVariant == "istem" || InputVerb.ConjugationVariant == "eyk") && InputVerb.Conjugation == 4 {
		Namespace = "comb" // all of the above variants use the same namespace
	} else {
		Namespace = InputVerb.ConjugationVariant // all other variants use their own variant string as a namespace
	}

	// present affirmative
	FormIndex = fmt.Sprintf("%d.pres.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	present, readErr := readForms(InputVerb.Stem, FormIndex)                // read the forms in that object
	if readErr != nil {                                                     // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	OutputArray = append(OutputArray, present)

	// present negative
	FormIndex = fmt.Sprintf("%d.pres.neg.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.Stem, FormIndex)              // read the forms in that object
	if readErr != nil {                                                         // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	presentNegative := negativeForms(temporaryForms) // make the forms negative
	OutputArray = append(OutputArray, presentNegative)

	// past direct affirmative
	FormIndex = fmt.Sprintf("%d.past.dir.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	pastDirect, readErr := readForms(InputVerb.Stem, FormIndex)                 // read the forms in that object
	if readErr != nil {                                                         // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	OutputArray = append(OutputArray, pastDirect)

	// past direct negative
	FormIndex = fmt.Sprintf("%d.past.dir.neg.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.Stem, FormIndex)                  // read the forms in that object
	if readErr != nil {                                                             // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	pastDirectNegative := negativeForms(temporaryForms) // make the forms negative
	OutputArray = append(OutputArray, pastDirectNegative)

	// past suppositive affirmative
	FormIndex = fmt.Sprintf("%d.past.sup.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	pastSuppositive, readErr := readForms(InputVerb.Stem, FormIndex)            // read the forms in that object
	if readErr != nil {                                                         // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	OutputArray = append(OutputArray, pastSuppositive)

	// past suppositive negative
	FormIndex = fmt.Sprintf("%d.past.sup.neg.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.Stem, FormIndex)                  // read the forms in that object
	if readErr != nil {                                                             // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	pastSuppositiveNegative := negativeForms(temporaryForms) // make the forms negative
	OutputArray = append(OutputArray, pastSuppositiveNegative)

	// past deferential affirmative
	FormIndex = fmt.Sprintf("%d.past.def.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	pastDeferential, readErr := readForms(InputVerb.Stem, FormIndex)            // read the forms in that object
	if readErr != nil {                                                         // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	OutputArray = append(OutputArray, pastDeferential)

	// past deferential negative
	FormIndex = fmt.Sprintf("%d.past.def.neg.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.Stem, FormIndex)                  // read the forms in that object
	if readErr != nil {                                                             // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	pastDeferentialNegative := negativeForms(temporaryForms) // make the forms negative
	OutputArray = append(OutputArray, pastDeferentialNegative)

	// future affirmative
	FormIndex = fmt.Sprintf("%d.futr.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	future, readErr := readForms(InputVerb.ContractedStem, FormIndex)       // read the forms in that object
	if readErr != nil {                                                     // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	OutputArray = append(OutputArray, future)

	// future negative
	FormIndex = fmt.Sprintf("%d.pres.neg.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.ContractedStem, FormIndex)    // read the forms in that object
	if readErr != nil {                                                         // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	var futureNegative []string
	if InputVerb.Conjugation == 6 || InputVerb.Conjugation == 7 {
		futureNegative = futureNegativeFormsVTA(temporaryForms) // VTA negatives are handled differently because of the separators
	} else {
		futureNegative = futureNegativeForms(temporaryForms) // make the forms negative
	}
	OutputArray = append(OutputArray, futureNegative)

	// imperative affirmative
	FormIndex = fmt.Sprintf("%d.impe.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	imperative, readErr := readForms(InputVerb.ContractedStem, FormIndex)   // read the forms in that object
	if readErr != nil {                                                     // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	OutputArray = append(OutputArray, imperative)

	// imperative negative
	FormIndex = fmt.Sprintf("%d.impe.neg.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.ContractedStem, FormIndex)    // read the forms in that object
	if readErr != nil {                                                         // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	var imperativeNegative []string
	if InputVerb.Conjugation == 6 || InputVerb.Conjugation == 7 {
		imperativeNegative = imperativeNegativeFormsVTA(temporaryForms) // VTA negatives are handled differently because of the separators
	} else {
		imperativeNegative = imperativeNegativeForms(temporaryForms) // make the forms negative
	}
	OutputArray = append(OutputArray, imperativeNegative)

	// when conjunct affirmative
	FormIndex = fmt.Sprintf("%d.when.prs.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	whenConjunct, readErr := readForms(InputVerb.Stem, FormIndex)               // read the forms in that object
	if readErr != nil {                                                         // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	OutputArray = append(OutputArray, whenConjunct)

	// when conjunct negative
	FormIndex = fmt.Sprintf("%d.when.prs.neg.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.Stem, FormIndex)                  // read the forms in that object
	if readErr != nil {                                                             // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	whenConjunctNegative := negativeForms(temporaryForms) // make the forms negative
	OutputArray = append(OutputArray, whenConjunctNegative)

	// when conjunct past
	FormIndex = fmt.Sprintf("%d.when.pst.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	whenConjunctPast, readErr := readForms(InputVerb.Stem, FormIndex)           // read the forms in that object
	if readErr != nil {                                                         // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	OutputArray = append(OutputArray, whenConjunctPast)

	// when conjunct past negative
	FormIndex = fmt.Sprintf("%d.when.pst.neg.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.Stem, FormIndex)                  // read the forms in that object
	if readErr != nil {                                                             // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	whenConjunctPastNegative := negativeForms(temporaryForms) // make the forms negative
	OutputArray = append(OutputArray, whenConjunctPastNegative)

	// if conjunct
	FormIndex = fmt.Sprintf("%d.when.prs.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.ContractedStem, FormIndex)    // read the forms in that object
	if readErr != nil {                                                         // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	// the if conjunct is the same as the when conjunct in the present, but with a contracted stem
	var ifConjunct []string
	if InputVerb.Conjugation == 6 || InputVerb.Conjugation == 7 { // no forms need to be removed for the sixth and seventh conjugations
		ifConjunct = temporaryForms
	} else {
		for formIndex, form := range temporaryForms { // some forms need to be removed from the when conjunct for up to the fifth conjugation
			if formIndex != 4 && formIndex != 5 && formIndex != 12 &&
				formIndex != 13 && formIndex != 19 && formIndex != 20 {
				ifConjunct = append(ifConjunct, form)
			}
		}
	}
	OutputArray = append(OutputArray, ifConjunct)

	// if conjunct negative
	FormIndex = fmt.Sprintf("%d.when.prs.neg.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.ContractedStem, FormIndex)        // read the forms in that object
	if readErr != nil {                                                             // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	// the if conjunct is the same as the when conjunct in the present, but with a contracted stem
	var ifConjunctNegative []string
	if InputVerb.Conjugation == 6 || InputVerb.Conjugation == 7 { // no forms need to be removed for the sixth and seventh conjugations
		ifConjunctNegative = negativeForms(temporaryForms) // make the forms negative
	} else {
		for formIndex, form := range temporaryForms { // some forms need to be removed from the when conjunct for up to the fifth conjugation
			if formIndex != 4 && formIndex != 5 && formIndex != 12 &&
				formIndex != 13 && formIndex != 19 && formIndex != 20 {
				ifConjunctNegative = append(ifConjunctNegative, form)
			}
		}
		ifConjunctNegative = negativeForms(ifConjunctNegative) // make the forms negative
	}
	OutputArray = append(OutputArray, ifConjunctNegative)

	// if conjunct suppositive
	FormIndex = fmt.Sprintf("%d.ifcn.sup.%s", InputVerb.Conjugation, Namespace)      // create the indexed title key
	ifConjunctSuppositive, readErr := readForms(InputVerb.ContractedStem, FormIndex) // read the forms in that object
	if readErr != nil {                                                              // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	OutputArray = append(OutputArray, ifConjunctSuppositive)

	// if conjunct suppositive negative
	FormIndex = fmt.Sprintf("%d.ifcn.sup.neg.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.Stem, FormIndex)                  // read the forms in that object
	if readErr != nil {                                                             // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	ifConjunctSuppositiveNegative := negativeForms(temporaryForms) // make the forms negative
	OutputArray = append(OutputArray, ifConjunctSuppositiveNegative)

	// if conjunct counterfactual
	FormIndex = fmt.Sprintf("%d.ifcn.cfl.%s", InputVerb.Conjugation, Namespace)         // create the indexed title key
	ifConjunctCounterfactual, readErr := readForms(InputVerb.ContractedStem, FormIndex) // read the forms in that object
	if readErr != nil {                                                                 // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	OutputArray = append(OutputArray, ifConjunctCounterfactual)

	// if conjunct counterfactual negative
	FormIndex = fmt.Sprintf("%d.ifcn.cfl.neg.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.Stem, FormIndex)                  // read the forms in that object
	if readErr != nil {                                                             // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	ifConjunctCounterfactualNegative := negativeForms(temporaryForms) // make the forms negative
	OutputArray = append(OutputArray, ifConjunctCounterfactualNegative)

	// conditional affirmative
	FormIndex = fmt.Sprintf("%d.cond.prs.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	conditional, readErr := readForms(InputVerb.ContractedStem, FormIndex)      // read the forms in that object
	if readErr != nil {                                                         // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	OutputArray = append(OutputArray, conditional)

	// some forms of the conditional do not have negatives in use; the future negative is used instead
	// conditional suppositive
	if InputVerb.Conjugation != 6 && InputVerb.Conjugation != 7 && InputVerb.Type != VII {
		// the sixth, seventh, and VII verbs do not have the suppositive conditional
		// (does not exist for inanimates, apparently conflated with the counterfactual in the 6th and 7th conjugations)
		FormIndex = fmt.Sprintf("%d.cond.sup.%s", InputVerb.Conjugation, Namespace)       // create the indexed title key
		conditionalSuppositive, readErr := readForms(InputVerb.ContractedStem, FormIndex) // read the forms in that object
		if readErr != nil {                                                               // if the forms are not read, the function will return an error
			fmt.Println(readErr)
		}
		OutputArray = append(OutputArray, conditionalSuppositive)
	}

	// conditional counterfactual
	FormIndex = fmt.Sprintf("%d.cond.cfl.%s", InputVerb.Conjugation, Namespace)          // create the indexed title key
	conditionalCounterfactual, readErr := readForms(InputVerb.ContractedStem, FormIndex) // read the forms in that object
	if readErr != nil {                                                                  // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	OutputArray = append(OutputArray, conditionalCounterfactual)

	// conditional counterfactual negative
	FormIndex = fmt.Sprintf("%d.cond.cfl.neg.%s", InputVerb.Conjugation, Namespace) // create the indexed title key
	temporaryForms, readErr = readForms(InputVerb.ContractedStem, FormIndex)        // read the forms in that object
	if readErr != nil {                                                             // if the forms are not read, the function will return an error
		fmt.Println(readErr)
	}
	conditionalCounterfactualNegative := negativeForms(temporaryForms) // make the forms negative
	OutputArray = append(OutputArray, conditionalCounterfactualNegative)

	// make the plural forms for VTI verbs
	if (InputVerb.Conjugation == 4 || InputVerb.Conjugation == 5) && InputVerb.ConjugationVariant != "inan" && InputVerb.ConjugationVariant != "eyk" {
		OutputArray = pluralInanimateForms(OutputArray)
	}

	return OutputArray
}

// this function will read the forms from the conjugation dictionary (loaded from conjdict.json)
// called by readoutVerb
func readForms(ToConcatenate string, FormIndex string) ([]string, error) {
	var FoundForms []string  // a slice of forms found in the file
	var OutputForms []string // the slice of completed (appended) forms
	var ResultForm string    // the resulting form of one append instance
	for _, element := range ConjugationDictionary {
		if element.Title == FormIndex { // iterate until the Title is equal to the FormIndex (i.e. the correct list in the file)
			FoundForms = element.Forms
		}
	}
	if len(FoundForms) == 0 { // if FoundForms is 0 long, the forms were not found
		return FoundForms, errors.New("Could not find forms in dictionary.")
	} else { // otherwise, the forms have been found
		for _, line := range FoundForms {
			if strings.Contains(line, "*") == true || strings.Contains(line, "&&") {
				// starred forms do not exist, and thus shouldn't have the stem prepended
				// "&&" is a delineator character between persons in VTA verbs
				ResultForm = line
			} else if strings.Contains(line, ":") == true { // forms separated by a colon are variants of a form
				// they should be concatenated together with a comma
				SplitForms := strings.Split(line, ":")
				ResultForm1 := SplitForms[0]
				ResultForm2 := SplitForms[1]
				ResultForm = fmt.Sprintf("%s%s, %s%s", ToConcatenate, ResultForm1, ToConcatenate, ResultForm2)
			} else {
				ResultForm = fmt.Sprintf("%s%s", ToConcatenate, line) // otherwise, simply append the form to the stem
			}
			OutputForms = append(OutputForms, ResultForm) // append the forms into a slice
		}
		return OutputForms, nil
	}
}

// generate the VTI plural agreement forms procedurally
func pluralInanimateForms(InputArray [][]string) [][]string {
	// only the present and past show agreement with plural inanimate objects
	// Pacifique also mentions 3rd person subjects in the future, but these forms appear old-fashioned
	var OutputArray [][]string = InputArray // the output composite literal
	var pluralForm string                   // the resulting plural form for any given form

	for tense := 0; tense < 2; tense++ { // the present
		currentTense := InputArray[tense]           // should be a slice of []string for the current tense
		outputTense := append(currentTense, "||")   // "||" is a delineator for VTI verbs. i could just as easily have used "&&" again.
		for formIndex, form := range currentTense { // for all forms in the current tense
			if form == "*" {
				pluralForm = "*"
			} else {
				if formIndex == 0 {
					if strings.Contains(form, ",") == true { // if there is a comma, i.e. if there are multiple variants
						splitForms := strings.Split(form, ", ")                                // split the string at that comma
						pluralForm = fmt.Sprintf("%sanl, %sanl", splitForms[0], splitForms[1]) // put "anl" after both variants, and join them together again
					} else {
						pluralForm = fmt.Sprintf("%sanl", form) // the first person singular takes -anl
					}
				} else {
					if string(form[len(form)-1]) == "l" { // if a form ends in -l, the plural is the same
						pluralForm = form
					} else {
						if strings.Contains(form, ",") == true { // if there is a comma, i.e. if there are multiple variants
							splitForms := strings.Split(form, ", ")                            // split the string at that comma
							pluralForm = fmt.Sprintf("%sl, %sl", splitForms[0], splitForms[1]) // put "l" after both variants, and join them together again
						} else {
							pluralForm = fmt.Sprintf("%sl", form) // else, append -l
						}
					}
				}
			}
			outputTense = append(outputTense, pluralForm) // append the plural form to the output tense
		}
		OutputArray[tense] = outputTense // set output tenses to override the original array tenses
		outputTense = nil                // blank out this slice
	}
	for tense := 2; tense < 8; tense++ { // the past
		currentTense := InputArray[tense]         // should be a slice of []string for the current tense
		outputTense := append(currentTense, "||") // "||" is a delineator for VTI verbs. i could just as easily have used "&&" again.
		for _, form := range currentTense {       // for all forms in the current tense
			if form == "*" {
				pluralForm = "*"
			} else {
				if strings.Contains(form, ",") {
					splitForms := strings.Split(form, ", ")
					pluralForm = fmt.Sprintf("%sl, %sl", splitForms[0], splitForms[1]) // the first person singular takes -anl
				} else {
					if string(form[len(form)-1]) == "n" || string(form[len(form)-1]) == "k" { // if a form ends in -n or -k
						pluralForm = fmt.Sprintf("%sl", form)
					} else if string(form[len(form)-1]) == "l" { // if the form ends in -l, the plural is the same
						pluralForm = form
					} else {
						pluralForm = fmt.Sprintf("%snl", form) // else, append -nl (for the past, like teluisiyekɨp-nl)
					}
				}
			}
			outputTense = append(outputTense, pluralForm) // append the plural form to the output tense
		}
		OutputArray[tense] = outputTense // set output tenses to override the original array tenses
		outputTense = nil                // blank out this slice
	}
	return OutputArray
}

// returns the plain negative forms of the tenses that use them (present negative, past negatives, etc.) — just "mu ..."
func negativeForms(InputForms []string) []string { //returns the negative form of a verb (essentially prepends "mu")
	var OutputForms []string          // output slice
	for _, form := range InputForms { // for every form in the input slice
		if form != "*" && form != "||" && form != "&&" { // these are delineator characters and should not be made negative
			if strings.Contains(form, ",") == true { // if there is a comma, i.e. if there are multiple variants
				splitForms := strings.Split(form, ", ")                                  // split the string at that comma
				joinedForms := fmt.Sprintf("mu %s, mu %s", splitForms[0], splitForms[1]) // put "mu" before both variants, and join them together again
				OutputForms = append(OutputForms, joinedForms)                           // append the result to the slice
			} else {
				form = fmt.Sprintf("mu %s", form) // otherwise, just add "mu"
				OutputForms = append(OutputForms, form)
			}
		} else {
			OutputForms = append(OutputForms, form)
		}
	}
	return OutputForms // return the final list
}

// returns the future negative (ma' ...)
func futureNegativeForms(InputForms []string) []string { // the negative for the future is "ma'"
	var OutputForms []string                  // output slice
	for formIndex, form := range InputForms { // for all forms in the input slice
		if strings.Contains(form, ",") == true { // if there are variant forms with a comma
			splitStrings := strings.Split(form, ", ")                              // split the string at the comma
			form = fmt.Sprintf("ma' %s, ma' %s", splitStrings[0], splitStrings[1]) // prepend "ma'" to each
		} else {
			form = fmt.Sprintf("ma' %s", form) // prepend "ma'"
		}
		// since the future negative is the same as the present negative, with ma' instead of mu
		// we can use the same slice, but we have to remove some persons
		if formIndex != 5 && formIndex != 6 &&
			formIndex != 14 && formIndex != 15 &&
			formIndex != 22 && formIndex != 23 {
			OutputForms = append(OutputForms, form) // append all persons that appear in the future (i.e. not the absentatives)
		}
	}
	return OutputForms // return the final slice
}

// the VTA verbs need a separate future negative function because the forms do not need to be filtered (missing forms are instead "*" in VTA lists)
func futureNegativeFormsVTA(InputForms []string) []string { // VTA verbs have escape characters, and all persons are used
	var OutputForms []string          // output slice
	for _, form := range InputForms { // for all forms in the input list
		if form != "*" && form != "&&" { // if not a delineator character
			form = fmt.Sprintf("ma' %s", form) // prepend ma'
		}
		OutputForms = append(OutputForms, form) // append the result
	}
	return OutputForms // return the final slice
}

// this returns the imperative negative ("mukk ..." for the second persons, "mu ..." otherwise)
func imperativeNegativeForms(InputForms []string) []string { // for the imperative negative forms (2nd persons take mukk, not mu)
	var OutputForms []string                  // output slice
	for formIndex, form := range InputForms { // for all forms in the input list
		if formIndex == 0 || formIndex == 5 || formIndex == 9 { // if the index corresponds to a second person
			if strings.Contains(form, ",") == true { // if there are variants of a form (with commas in between)
				splitStrings := strings.Split(form, ", ")                                // split the string at the comma
				form = fmt.Sprintf("mukk %s, mukk %s", splitStrings[0], splitStrings[1]) // prepend "mukk" to each
			} else {
				form = fmt.Sprintf("mukk %s", form) // prepend "mukk"
			}
		} else {
			if strings.Contains(form, ",") == true { // if there are variants with a comma
				splitStrings := strings.Split(form, ", ") // split the string at the comma
				form = fmt.Sprintf("mu %s, mu %s", splitStrings[0], splitStrings[1])
			} else {
				form = fmt.Sprintf("mu %s", form) // else, prepend "mu"
			}
		}
		OutputForms = append(OutputForms, form) // append the resulting form to the slice
	}
	return OutputForms // return the slice
}

// the imperative negative forms of VTA verbs need a separate function because the person index that corresponds to the second persons is different, and the forms need to be split at "&&" for that to work
func imperativeNegativeFormsVTA(InputForms []string) []string { // VTA verbs work differently here because there are multiple indices that correspond to second persons
	var OutputForms []string                          // the output slice
	var splitForms2 [][]string                        // a composite literal to hold forms split twice
	joinedForms := strings.Join(InputForms, "%")      // join all forms in the input slice with "%"
	splitForms1 := strings.Split(joinedForms, "%&&%") // split all forms in the input at "%&&%". this returns a string for each person
	for _, form := range splitForms1 {                // for every form in the split list
		formSlice := strings.Split(form, "%")        // split them again at "%"
		splitForms2 = append(splitForms2, formSlice) // append the slice for each person to split forms2
	}
	for _, form := range splitForms2 { // for every person slice
		var temporaryForms []string         // a temporary slice for doing joining/concatenation
		for itemIndex, item := range form { // for every item in that person slice
			if item != "*" { // if it is not a delineator character
				if itemIndex == 0 || itemIndex == 5 || itemIndex == 8 { // if the index is 0 or 5 or 8 (corresponding to 2nd persons for each person slice)
					if strings.Contains(item, ",") == true { // if there are comma separated variant forms
						formSlice := strings.Split(item, ", ")                             // split the string
						item = fmt.Sprintf("mukk %s, mukk %s", formSlice[0], formSlice[1]) // prepend "mukk" to both
					} else {
						item = fmt.Sprintf("mukk %s", item) // prepend mukk
					}
				} else {
					if strings.Contains(item, ",") == true { // if there are comma separated variant forms
						formSlice := strings.Split(item, ", ")                         // split the string there
						item = fmt.Sprintf("mu %s, mu %s", formSlice[0], formSlice[1]) // prepend "mu" to both
					} else {
						item = fmt.Sprintf("mu %s", item) // else, prepend mu
					}
				}
			}
			temporaryForms = append(temporaryForms, item) // append each item to a temporary slice
		}
		temporaryForms = append(temporaryForms, "&&") // between each person slice being appended to the temporary slice, insert the delineator "&&"
		OutputForms = append(OutputForms, temporaryForms...)
	}
	OutputForms = OutputForms[:len(OutputForms)-1] // remove the final string from the output slice (this is an extra "&&")
	return OutputForms                             // return the output slice
}

// this function will create a type Verb by recognizing the group of the input stem
// called by readoutVerb
func parseVerb(InputStr string) (Verb, error) {
	// replace ɨ with *
	if strings.Contains(InputStr, "ɨ") == true {
		InputStr = strings.Replace(InputStr, "ɨ", "*", -1)
	}

	var InputVerb Verb // define an instance of the Verb struct
	var Ending string  // for storing the ending of the InputStr
	var FinalInt int   // for how many characters you are looking at the end of a verb

	// exceptions
	// i had thought about putting e.g. "etek" here, as an exception to "eyk", etc.,
	// but i think these are rather separate verbs, and the animate and inanimate conjugations are not combined

	// start with 4 letter verb endings
	FinalInt = 4
	if len(InputStr) > 3 { // strings with length less than 4 will throw an error
		Ending = getVerbEnding(InputStr, FinalInt)
	}
	if Ending == "a'tl" {
		// there are two possibilities here: -a stem and -e stem verbs both of which end in -a'tl for the third person
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 6
		InputVerb.ConjugationVariant = "aestem"
		// InputVerb.ConjugationVariant = "astem"
		// InputVerb.ConjugationVariant = "estem"
		// the above are relegated to the python script — had to do things a bit differently here and so the above lines are commented out
		InputVerb.Type = VTA
		return InputVerb, nil
	} else if Ending == "a's*k" { // first conjugation inanimate verbs in "a'sɨk" (orthographical/Listuguj variant of "a'sik")
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 1
		InputVerb.ConjugationVariant = "asik"
		InputVerb.Type = VII
		return InputVerb, nil
	}

	// now do 3 letter verb endings
	FinalInt = 3
	if len(InputStr) > 2 { // strings with length less than 3 will throw an error
		Ending = getVerbEnding(InputStr, FinalInt)
	}
	if Ending == "a't" { // second conjugation verbs with a long vowel
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 2
		InputVerb.ConjugationVariant = "long"
		InputVerb.Type = VAI
		return InputVerb, nil
	} else if Ending == "ayk" { // conjugation 1~2 verbs with a diphthong
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 2
		InputVerb.ConjugationVariant = "diph"
		InputVerb.Type = VAI
		return InputVerb, nil
	} else if Ending == "iaq" { // third conjugation inanimate verbs that resemble -iet's inanimate conjugation
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 3
		InputVerb.ConjugationVariant = "iaq"
		InputVerb.Type = VII
		return InputVerb, nil
	} else if Ending == "e'k" {
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		FinalInt = 4
		Ending = getVerbEnding(InputStr, FinalInt)
		if Ending == "te'k" { // fourth conjugation verbs in -te'k, e.g. telte'k
			FinalInt = 1
			InputVerb.Stem = getVerbStem(InputStr, FinalInt)
			InputVerb.Conjugation = 4
			InputVerb.ConjugationVariant = "estem"
			InputVerb.Type = VTI
			return InputVerb, nil
		} else { // third conjugation verbs with a long vowel, also fit somewhat in the first conjugation
			FinalInt = 3
			InputVerb.Stem = getVerbStem(InputStr, FinalInt)
			InputVerb.Conjugation = 3
			InputVerb.ConjugationVariant = "long"
			InputVerb.Type = VAI
			return InputVerb, nil
		}
	} else if Ending == "ink" { // conjugation 1~4 verbs in -ink
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 1
		InputVerb.ConjugationVariant = "ink"
		InputVerb.Type = VAI
		return InputVerb, nil
	} else if Ending == "t*k" { // fourth conjugation verbs that end in An -t, with an intervening schwa
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt+1) // ɨ counts as two characters?
		InputVerb.Conjugation = 4
		InputVerb.ConjugationVariant = "ibar"
		InputVerb.Type = VTI
		return InputVerb, nil
	} else if Ending == "a'q" { // fourth conjugation verbs in -a'q, overlaps some second person inanimate verbs, but those are covered with -a't
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 4
		InputVerb.ConjugationVariant = "astem"
		InputVerb.Type = VTI
		return InputVerb, nil
	} else if Ending == "i'k" { // fourth conjugation verbs with an -i- stem
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		FinalInt = 1 // have to only remove the -k
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 4
		InputVerb.ConjugationVariant = "istem"
		InputVerb.Type = VTI
		return InputVerb, nil
	} else if Ending == "toq" { // fifth conjugation verbs in -toq, e.g. muska'toq
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		FinalInt = 2
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 5
		InputVerb.ConjugationVariant = "std"
		InputVerb.Type = VTI
		return InputVerb, nil
	} else if Ending == "atl" { // variations of VTA verbs
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Type = VTA
		FinalInt = 4
		Ending = getVerbEnding(InputStr, FinalInt)
		if Ending == "iatl" { // i.e. -iatl (e.g. nemiatl)
			InputVerb.Stem = getVerbStem(InputStr, FinalInt)
			InputVerb.Conjugation = 6
			InputVerb.ConjugationVariant = "istem"
			return InputVerb, nil
		} else if Ending == "uatl" { // i.e. "mixed" VTA verbs, -uatl (e.g. kwiluatl)
			InputVerb.Stem = getVerbStem(InputStr, FinalInt)
			InputVerb.Conjugation = 7
			InputVerb.ConjugationVariant = "std"
			return InputVerb, nil
		} else { // other sixth conjugation verbs in -atl, e.g. kesalatl
			FinalInt = 3
			InputVerb.Stem = getVerbStem(InputStr, FinalInt)
			InputVerb.Conjugation = 6
			// check if the verb has a long vowel + consonant before the personal agreement ending, e.g. e'natl
			// these verbs require a schwa inserted for phonotactic reasons
			if IsConsonant(string(InputVerb.Stem[len(InputVerb.Stem)-1])) == true && string(InputVerb.Stem[len(InputVerb.Stem)-2]) == "'" {
				InputVerb.ConjugationVariant = "ibar"
			} else {
				InputVerb.ConjugationVariant = "std"
			}
			return InputVerb, nil
		}
	} else if Ending == "u'k" { // fourth conjugation verbs with inanimate subject, e.g. telamu'k
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 4
		InputVerb.ConjugationVariant = "inan"
		InputVerb.Type = VII
		return InputVerb, nil
	}

	// then do 2 letter verb endings
	FinalInt = 2
	if len(InputStr) > 1 { // strings with length 1 will throw an error
		Ending = getVerbEnding(InputStr, FinalInt)
	}
	if Ending == "it" { // Pacifique's first conjugation
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Conjugation = 1
		InputVerb.Type = VAI
		FinalInt = 5 // need to check for verbs ending in "a'sit"
		Ending = getVerbEnding(InputStr, FinalInt)
		if Ending == "a'sit" {
			InputVerb.Stem = getVerbStem(InputStr, FinalInt)
			InputVerb.ConjugationVariant = "asit"
		} else {
			FinalInt = 2
			InputVerb.Stem = getVerbStem(InputStr, FinalInt)
			InputVerb.ConjugationVariant = "std"
		}
		return InputVerb, nil
	} else if Ending == "at" { // Pacifique's second conjugation
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 2
		InputVerb.ConjugationVariant = "std"
		InputVerb.Type = VAI
		return InputVerb, nil
	} else if Ending == "et" { // Pacifique's third conjugation
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Conjugation = 3
		InputVerb.Type = VAI
		FinalInt = 3 // start checking for variants
		Ending = getVerbEnding(InputStr, FinalInt)
		if Ending == "iet" { // third conjugation verbs in -iet
			InputVerb.Stem = getVerbStem(InputStr, FinalInt)
			InputVerb.ConjugationVariant = "iet"
		} else if Ending == "uet" || Ending == "wet" { // third conjugation verbs in -uet/wet
			FinalInt = 2 // set this back to 2 to keep the -w or -u of the stem
			InputVerb.Stem = getVerbStem(InputStr, FinalInt)
			InputVerb.ConjugationVariant = "uet"
		} else { // check variants beyond 3 final characters
			FinalInt = 4
			Ending = getVerbEnding(InputStr, FinalInt)
			if Ending == "eket" { // third conjugation verbs in -eket
				InputVerb.Stem = getVerbStem(InputStr, FinalInt)
				InputVerb.ConjugationVariant = "eket"
			} else { // if not one of the above variants, it is a standard third conjugation verb
				FinalInt = 2
				InputVerb.Stem = getVerbStem(InputStr, FinalInt)
				InputVerb.ConjugationVariant = "std"
			}
		}
		return InputVerb, nil
	} else if Ending == "tk" { // some of Pacifique's fourth conjugation
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 4
		InputVerb.ConjugationVariant = "std"
		InputVerb.Type = VTI
		return InputVerb, nil
	} else if string(Ending[1]) == "k" && IsConsonant(string(Ending[0])) == true && string(Ending[0]) != "t" { // more of Pacifique's fourth conjugation
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		FinalInt = 1
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 4
		if InputStr == "eyk" { // eyk acts like an intransitive verb
			InputVerb.Type = VAI
			InputVerb.ConjugationVariant = "eyk"
		} else {
			InputVerb.Type = VTI
			InputVerb.ConjugationVariant = "cons"
		}
		return InputVerb, nil
	} else if Ending == "*k" && string(InputStr[len(InputStr)-1:]) != "t" { // fourth conjugation verbs with stems in -kɨk
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt+1) // ɨ counts as two characters?
		InputVerb.Conjugation = 4
		InputVerb.ConjugationVariant = "kstem"
		InputVerb.Type = VTI
		return InputVerb, nil
	} else if Ending == "uk" { // fifth conjugation verbs that do not take "oq" in the third person
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 5
		InputVerb.ConjugationVariant = "kuk"
		InputVerb.Type = VTI
		return InputVerb, nil
	} else if Ending == "ik" { // first conjugation verbs with inanimate subjects
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 1
		FinalInt = 5 // need to check for verbs ending in "a'sik"
		Ending = getVerbEnding(InputStr, FinalInt)
		if Ending == "a'sik" {
			InputVerb.Stem = getVerbStem(InputStr, FinalInt)
			InputVerb.ConjugationVariant = "asik"
		} else {
			FinalInt = 2
			InputVerb.Stem = getVerbStem(InputStr, FinalInt)
			InputVerb.ConjugationVariant = "inan"
		}
		return InputVerb, nil
	} else if Ending == "aq" { // second conjugation verbs with inanimate subjects
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 2
		InputVerb.ConjugationVariant = "inan"
		InputVerb.Type = VII
		return InputVerb, nil
	} else if Ending == "ek" { // third conjugation verbs with inanimate subjects
		if strings.Contains(InputStr, "*") == true { // convert the stars back to ɨ
			InputStr = strings.Replace(InputStr, "*", "ɨ", -1)
		}
		InputVerb.Stem = getVerbStem(InputStr, FinalInt)
		InputVerb.Conjugation = 3
		InputVerb.ConjugationVariant = "inan"
		InputVerb.Type = VII
		return InputVerb, nil
	}

	return InputVerb, errors.New("Verb Unrecognized")
}

// this returns the input string minus the last FinalInt characters (i.e. the stem)
func getVerbStem(InputStr string, FinalInt int) string {
	var OutputStr string
	OutputStr = InputStr[:len(InputStr)-FinalInt] // gives InputStr minues the last FinalInt characters
	return OutputStr
}

// this returns the last FinalInt characters of the input string (i.e. the ending of the verb)
func getVerbEnding(InputStr string, FinalInt int) string {
	var OutputStr string
	OutputStr = InputStr[len(InputStr)-FinalInt:] // gives the last FinalInt characters of InputStr
	return OutputStr
}

// this returns a contracted stem — verbs with "e" in the first syllable have it removed, and there are different phonotactic consequences for this
func contractStem(InputStr string, Conjugation int) string { // return the contracted stem for use in the future, etc.
	var OutputStr string
	if string(InputStr[0]) == "e" && IsConsonant(string(InputStr[1])) == true {
		OutputStr = InputStr[1:]
		if string(OutputStr[0]) == "y" { // if the first character is y, turn this into i'. e.g. ey- => y- => i'-
			OutputStr = fmt.Sprintf("i'%s", OutputStr[1:])
		}
	} else if string(InputStr[1]) == "e" && IsConsonant(string(InputStr[2])) { // if the second character is e and the third character is a consonant
		// should not matter if the first character is a consonant or vowel, since if e is the second character, the first should be a consonant anyways
		OutputStr = fmt.Sprintf("%s%s", string(InputStr[0]), InputStr[2:])
		if string(OutputStr[0]) == "y" { // see above; if the first character is y, turn this into i'. e.g. ey- => y- => i'-
			OutputStr = fmt.Sprintf("i'%s", OutputStr[1:])
		}
		if len(OutputStr) > 2 { //handle stems shorter than two differently
			if IsConsonant(string(OutputStr[0])) == true && IsPlosive(string(OutputStr[1])) == true && IsPlosive(string(OutputStr[2])) == true {
				// if the first character is a consonant, the second is a plosive, and the third is a plosive
				if IsPlosive(string(OutputStr[0])) == true { // if the first is a plosive
					OutputStr = fmt.Sprintf("%sɨ%s", OutputStr[:1], OutputStr[1:])
				} else { // if the first is not a plosive (therefore must be a sonorant, a consonant that is not a plosive)
					OutputStr = fmt.Sprintf("%sɨ%s", OutputStr[:2], OutputStr[2:])
				}
			}
		} else { // stems that are shorter than two characters
			if IsConsonant(string(OutputStr[0])) && IsPlosive(string(OutputStr[1])) && Conjugation != 6 && Conjugation != 7 {
				// if the first character is a consonant, and the second is a plosive
				// in VTA verbs (conj. 6, 7), the affixes contain vowels that do not necessitate the insertion of a schwa
				OutputStr = fmt.Sprintf("%sɨ", OutputStr)
			} else if OutputStr[0] == OutputStr[1] { // if the first and second characters are equal, e.g. for nen- => nn- => nɨn-
				OutputStr = fmt.Sprintf("%sɨ%s", string(OutputStr[0]), string(OutputStr[1]))
			}
		}
	} else {
		OutputStr = InputStr
	} // if no contraction can be made, the "contracted" stem is the same as the uncontracted one
	return OutputStr
}
