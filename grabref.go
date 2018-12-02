package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "regexp"
    "sort"
    "strconv"
    "strings"
)

type Reference struct {
    Year    int
    Authors []string
    Etal    bool
    Text    string
}

func main() {
    inputPath := flag.String("input", "", "path to the text file, use stdin if left empty")
    outputPath := flag.String("output", "", "path to the output file, stdout if left empty")
    sorted := flag.Bool("sorted", false, "sort authors lexicographically")
    dedup := flag.Bool("dedup", false, "deduplicate references, implies sorted")
    etal := flag.Bool("etal", false, "add 'et al.' to author list if it appears in input")
    //verbose := flag.Bool("verbose", false, "verbose output")
    //strictSearch := flag.Bool("strict", false, "strict Google Scholar search")
    flag.Parse()
    var err error

    /* open input file */
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

    /* regex for strings that we're pretty sure to be refs */
    accuRe, err := regexp.Compile(`(?m)(((([A-Z]\S+?)((, )|( and )|(, and )))*([A-Z]\S+?)( et al.?)? \(\d{4}\))|(\((e.g. )?((([A-Z]\S+?)((, )|( and )|(, and )))*([A-Z]\S+?)( et al.?)?,? (\d{4}); )*?((([A-Z]\S+?)((, )|( and )|(, and )))*([A-Z]\S+?)(,? et al.?)?,? (\d{4}))\)))`)
    /* regex matching four digits in a pair of parentheses */
    coarseRe, err := regexp.Compile(`(?m)\([^()]*?(\d{4})[^()]*?\)`)
    /* regex matching individual items in a set of refs separated by semicolons */
    subitemRe, err := regexp.Compile(`(?m)(([A-Z]\S+?)((, )|( and )|(, and )))*([A-Z]\S+?)(,? et al.?)?,? (\d{4})`)
    /* regex matching name, year, and semicolon */
    nameRe, err := regexp.Compile(`(?m)[A-Z]\p{L}+`)
    yearRe, err := regexp.Compile(`(?m)\d{4}`)
    semicolonRe, err := regexp.Compile(`(?m)\;`)
    etalRe, err := regexp.Compile(`(?m)et al`)

    text, err := ioutil.ReadAll(inF)

    if err != nil {
        log.Fatalf("Error reading input file")
    }

    confidentMatches := accuRe.FindAllIndex(text, -1)
    coarseMatches := coarseRe.FindAllIndex(text, -1)

    idxConfidentMatches := 0
    numConfidentMatches := len(confidentMatches)

    /* diff the results of coarse matching and fine matching, and report possible refs */
    for _, match := range coarseMatches {
        if (idxConfidentMatches >= numConfidentMatches) || (match[1] != confidentMatches[idxConfidentMatches][1]) {
            log.Println("Possible reference:", string(text[match[0]:match[1]]))
        } else {
            idxConfidentMatches += 1
        }
    }

    var references []Reference

    for _, match := range confidentMatches {
        /* decompose multiple refs separated by semicolons */
        switch semicolonRe.Match(text[match[0]:match[1]]) {
        case false:
            etal := etalRe.Match(text[match[0]:match[1]])
            names := nameRe.FindAllString(string(text[match[0]:match[1]]), -1)
            year, _ := strconv.Atoi(yearRe.FindString(string(text[match[0]:match[1]])))
            newRef := Reference{year, names, etal, string(text[match[0]:match[1]])}
            references = append(references, newRef)
        case true:
            for _, match := range subitemRe.FindAllString(string(text[match[0]:match[1]]), -1) {
                etal := etalRe.MatchString(match)
                names := nameRe.FindAllString(match, -1)
                year, _ := strconv.Atoi(yearRe.FindString(match))
                newRef := Reference{year, names, etal, match}
                references = append(references, newRef)
            }
        }
    }

    /* sort references lexicographically */
    compareRefs := func(i, j int) bool {
        authLen1 := len(references[i].Authors)
        authLen2 := len(references[j].Authors)
        min := 0
        if authLen1 < authLen2 {
            min = authLen1
        } else {
            min = authLen2
        }
        for a := 0; a < min; a++ {
            comp := strings.Compare(references[i].Authors[a], references[j].Authors[a])
            if comp < 0 {
                return true
            }
            if comp > 0 {
                return false
            }
        }
        if authLen1 < authLen2 {
            return true
        } else {
            return false
        }
    }

    if *sorted || *dedup {
        sort.SliceStable(references, compareRefs)
    }

    identicalRefs := func(i, j int) bool {
        if references[i].Year != references[j].Year {
            return false
        }
        if len(references[i].Authors) != len(references[j].Authors) {
            return false
        }
        refLen := len(references[i].Authors)
        for a := 0; a < refLen; a++ {
            if strings.Compare(references[i].Authors[a], references[j].Authors[a]) != 0 {
                return false
            }
        }
        return true
    }

    if *dedup {
        j := 0
        refLen := len(references)
        for i := 1; i < refLen; i++ {
            if identicalRefs(i, j) {
                fmt.Printf("[%v][%v]\n", references[i], references[j])
                continue
            }
            j++
            references[i], references[j] = references[j], references[i]
        }
        references = references[:j+1]
    }

    /* open output file */
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

    outF.WriteString(htmlHead)
    /* process each reference */
    for _, ref := range references {
        authorString := strings.Join(ref.Authors, ", ")
        if *etal && ref.Etal {
            authorString = strings.Join([]string{authorString, "et al."}, " ")
        }
        fmt.Fprintf(outF, htmlEntry,
        authorString,
        ref.Year,
        scholarQuery(&ref, true),
        scholarQuery(&ref, false),
        ref.Text)
    }
    outF.WriteString(`</ol></body></html>`)
}

func scholarQuery(ref *Reference, strict bool) string {
    nameList := strings.Join(ref.Authors, "+")
    year := strconv.Itoa(ref.Year)
    var query string
    if strict {
        query = strings.Join([]string{"https://scholar.google.com/scholar?as_q=&as_epq=&as_oq=&as_eq=&as_occt=any&as_sauthors=",
        nameList,
        "&as_publication=&as_ylo=",
        year,
        "&as_yhi=",
        year,
        "&hl=en&as_sdt=0%2C22"}, "")
    } else {
        query = strings.Join([]string{"https://scholar.google.com/scholar?hl=en&q=",
        year,
        "+",
        nameList}, "")
    }
    return query
}
