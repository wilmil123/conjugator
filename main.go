package main

import (
	"conjugator/bescherelle"
	"conjugator/converter"
	"fmt"
	"net/http"
	"text/template"
)

func main() {
	conjugatorErr := bescherelle.ConjugatorInit()
	if conjugatorErr != nil {
		fmt.Println(conjugatorErr)
	}

	converterErr := converter.ConverterInit()
	if converterErr != nil {
		fmt.Println(converterErr)
	}

	http.HandleFunc("/", mainIndexHandler) // create the webpage

	fileServe := http.FileServer(http.Dir("./assets"))              // add a stylesheet
	http.Handle("/assets/", http.StripPrefix("/assets", fileServe)) // no idea what this actually does, but this is from golang example code
	// all pages pull from the same stylesheet for consistency
	http.ListenAndServe(":8080", nil) // listen and serve
}

func mainIndexHandler(writer http.ResponseWriter, reader *http.Request) {
	template, templateBuildErr := template.ParseFiles("hometemplate.html.temp") // parse conjugatortemplate.html.temp
	if templateBuildErr != nil {                                                // if an error is thrown
		fmt.Println(templateBuildErr)
	}
	template.Execute(writer, nil) // execute the template
}
