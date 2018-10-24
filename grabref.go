package main

import (
    "flag"
    "os"
    "log"
    "io/ioutil"
//    "strings"
    "strconv"
//    "fmt"
    "regexp"
)

type Reference struct {
    Year    int
    Authors []string
    Text    string
}

func main() {
    inputPath := flag.String("input", "", "path to the text file, use stdin if left empty")
    outputPath := flag.String("output", "", "path to the output file, stdout if left empty")
    flag.Parse()
    var err error

    var inF *os.File
    defer inF.Close()
    if *inputPath != "" {
        inF, err = os.Open(*inputPath)
        if err != nil {
            log.Fatalf("Error opening input file at %s", *inputPath)
        }
    } else {
        inF = os.Stdin
    }

    accuRe, err := regexp.Compile(`(?m)(((([A-Z]\S+?)((, )|( and )|(, and )))*([A-Z]\S+?)( et al.?)? \(\d{4}\))|(\((e.g. )?((([A-Z]\S+?)((, )|( and )|(, and )))*([A-Z]\S+?)( et al.?)?,? (\d{4}); )*?((([A-Z]\S+?)((, )|( and )|(, and )))*([A-Z]\S+?)(,? et al.?)?,? (\d{4}))\)))`)
    coarseRe, err := regexp.Compile(`(?m)\([^()]*?(\d{4})[^()]*?\)`)
    subitemRe, err := regexp.Compile(`(?m)(([A-Z]\S+?)((, )|( and )|(, and )))*([A-Z]\S+?)(,? et al.?)?,? (\d{4})`)
    nameRe, err := regexp.Compile(`(?m)[A-Z]\p{L}+`)
    yearRe, err := regexp.Compile(`(?m)\d{4}`)
    semicolonRe, err := regexp.Compile(`(?m)\;`)

    text, err := ioutil.ReadAll(inF)

    if err != nil {
        log.Fatalf("Error reading input file")
    }

    confidentMatches := accuRe.FindAllIndex(text, -1)
    coarseMatches := coarseRe.FindAllIndex(text, -1)

    idxConfidentMatches := 0
    numConfidentMatches := len(confidentMatches)

    for _, match := range coarseMatches {
        if (idxConfidentMatches >= numConfidentMatches) || (match[1] != confidentMatches[idxConfidentMatches][1]) {
            log.Println("Possible reference:", string(text[match[0]:match[1]]))
        } else {
            idxConfidentMatches += 1
        }
    }

    var references []Reference

    for _, match := range confidentMatches {
        switch semicolonRe.Match(text[match[0]:match[1]]) {
        case false:
            names := nameRe.FindAllString(string(text[match[0]:match[1]]), -1)
            year, _ := strconv.Atoi(yearRe.FindString(string(text[match[0]:match[1]])))
            newRef := Reference{year, names, string(text[match[0]:match[1]])}
            references = append(references, newRef)
        case true:
            for _, match := range subitemRe.FindAllString(string(text[match[0]:match[1]]), -1) {
                names := nameRe.FindAllString(match, -1)
                year, _ := strconv.Atoi(yearRe.FindString(match))
                newRef := Reference{year, names, match}
                references = append(references, newRef)
            }
        }
    }

    var outF *os.File
    defer outF.Close()
    if *outputPath != "" {
        outF, err = os.Create(*outputPath)
        if err != nil {
            log.Fatalf("Error creating output file at %s", *outputPath)
        }
    } else {
        outF = os.Stdout
    }

    for _, ref := range references {
        outF.Write([]byte(strconv.Itoa(ref.Year)))
        for _, name := range ref.Authors {
            outF.Write([]byte(" "))
            outF.Write([]byte(name))
        }
        outF.Write([]byte("\n"))
    }
}
