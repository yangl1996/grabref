package main

import (
    "flag"
    "os"
    "log"
    "io/ioutil"
    "strings"
    "strconv"
    "fmt"
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

    outF.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>References</title>
<link rel="stylesheet" href="style.css">
<script>function copy(that){
var inp =document.createElement('input');
document.body.appendChild(inp)
inp.value =that.textContent
inp.select();
document.execCommand('copy',false);
inp.remove();
}</script>
<style>
body {
  font-family: Helvetica, arial, sans-serif;
  font-size: 14px;
  line-height: 1.6;
  padding-top: 10px;
  padding-bottom: 10px;
  background-color: white;
  padding: 30px; }

body > *:first-child {
  margin-top: 0 !important; }
body > *:last-child {
  margin-bottom: 0 !important; }

a {
  color: #4183C4; }
a.absent {
  color: #cc0000; }
a.anchor {
  display: block;
  padding-left: 30px;
  margin-left: -30px;
  cursor: pointer;
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0; }

h1, h2, h3, h4, h5, h6 {
  margin: 20px 0 10px;
  padding: 0;
  font-weight: bold;
  -webkit-font-smoothing: antialiased;
  cursor: text;
  position: relative; }

h1:hover a.anchor, h2:hover a.anchor, h3:hover a.anchor, h4:hover a.anchor, h5:hover a.anchor, h6:hover a.anchor {
  background: url("../../images/modules/styleguide/para.png") no-repeat 10px center;
  text-decoration: none; }

h1 tt, h1 code {
  font-size: inherit; }

h2 tt, h2 code {
  font-size: inherit; }

h3 tt, h3 code {
  font-size: inherit; }

h4 tt, h4 code {
  font-size: inherit; }

h5 tt, h5 code {
  font-size: inherit; }

h6 tt, h6 code {
  font-size: inherit; }

h1 {
  font-size: 28px;
  color: black; }

h2 {
  font-size: 24px;
  border-bottom: 1px solid #cccccc;
  color: black; }

h3 {
  font-size: 18px; }

h4 {
  font-size: 16px; }

h5 {
  font-size: 14px; }

h6 {
  color: #777777;
  font-size: 14px; }

p, blockquote, ul, ol, dl, li, table, pre {
  margin: 15px 0; }

hr {
  background: transparent url("../../images/modules/pulls/dirty-shade.png") repeat-x 0 0;
  border: 0 none;
  color: #cccccc;
  height: 4px;
  padding: 0; }

body > h2:first-child {
  margin-top: 0;
  padding-top: 0; }
body > h1:first-child {
  margin-top: 0;
  padding-top: 0; }
  body > h1:first-child + h2 {
    margin-top: 0;
    padding-top: 0; }
body > h3:first-child, body > h4:first-child, body > h5:first-child, body > h6:first-child {
  margin-top: 0;
  padding-top: 0; }

a:first-child h1, a:first-child h2, a:first-child h3, a:first-child h4, a:first-child h5, a:first-child h6 {
  margin-top: 0;
  padding-top: 0; }

h1 p, h2 p, h3 p, h4 p, h5 p, h6 p {
  margin-top: 0; }

li p.first {
  display: inline-block; }

ul, ol {
  padding-left: 30px; }

ul :first-child, ol :first-child {
  margin-top: 0; }

ul :last-child, ol :last-child {
  margin-bottom: 0; }

dl {
  padding: 0; }
  dl dt {
    font-size: 14px;
    font-weight: bold;
    font-style: italic;
    padding: 0;
    margin: 15px 0 5px; }
    dl dt:first-child {
      padding: 0; }
    dl dt > :first-child {
      margin-top: 0; }
    dl dt > :last-child {
      margin-bottom: 0; }
  dl dd {
    margin: 0 0 15px;
    padding: 0 15px; }
    dl dd > :first-child {
      margin-top: 0; }
    dl dd > :last-child {
      margin-bottom: 0; }

blockquote {
  border-left: 4px solid #dddddd;
  padding: 0 15px;
  color: #777777; }
  blockquote > :first-child {
    margin-top: 0; }
  blockquote > :last-child {
    margin-bottom: 0; }

table {
  padding: 0; }
  table tr {
    border-top: 1px solid #cccccc;
    background-color: white;
    margin: 0;
    padding: 0; }
    table tr:nth-child(2n) {
      background-color: #f8f8f8; }
    table tr th {
      font-weight: bold;
      border: 1px solid #cccccc;
      text-align: left;
      margin: 0;
      padding: 6px 13px; }
    table tr td {
      border: 1px solid #cccccc;
      text-align: left;
      margin: 0;
      padding: 6px 13px; }
    table tr th :first-child, table tr td :first-child {
      margin-top: 0; }
    table tr th :last-child, table tr td :last-child {
      margin-bottom: 0; }

img {
  max-width: 100%; }

span.frame {
  display: block;
  overflow: hidden; }
  span.frame > span {
    border: 1px solid #dddddd;
    display: block;
    float: left;
    overflow: hidden;
    margin: 13px 0 0;
    padding: 7px;
    width: auto; }
  span.frame span img {
    display: block;
    float: left; }
  span.frame span span {
    clear: both;
    color: #333333;
    display: block;
    padding: 5px 0 0; }
span.align-center {
  display: block;
  overflow: hidden;
  clear: both; }
  span.align-center > span {
    display: block;
    overflow: hidden;
    margin: 13px auto 0;
    text-align: center; }
  span.align-center span img {
    margin: 0 auto;
    text-align: center; }
span.align-right {
  display: block;
  overflow: hidden;
  clear: both; }
  span.align-right > span {
    display: block;
    overflow: hidden;
    margin: 13px 0 0;
    text-align: right; }
  span.align-right span img {
    margin: 0;
    text-align: right; }
span.float-left {
  display: block;
  margin-right: 13px;
  overflow: hidden;
  float: left; }
  span.float-left span {
    margin: 13px 0 0; }
span.float-right {
  display: block;
  margin-left: 13px;
  overflow: hidden;
  float: right; }
  span.float-right > span {
    display: block;
    overflow: hidden;
    margin: 13px auto 0;
    text-align: right; }

code, tt {
  margin: 0 2px;
  padding: 0 5px;
  white-space: nowrap;
  border: 1px solid #eaeaea;
  background-color: #f8f8f8;
  border-radius: 3px; }

pre code {
  margin: 0;
  padding: 0;
  white-space: pre;
  border: none;
  background: transparent; }

.highlight pre {
  background-color: #f8f8f8;
  border: 1px solid #cccccc;
  font-size: 13px;
  line-height: 19px;
  overflow: auto;
  padding: 6px 10px;
  border-radius: 3px; }

pre {
  background-color: #f8f8f8;
  border: 1px solid #cccccc;
  font-size: 13px;
  line-height: 19px;
  overflow: auto;
  padding: 6px 10px;
  border-radius: 3px; }
  pre code, pre tt {
    background-color: transparent;
    border: none; }
</style>
</head>
<body><ol>`)
    /* process each reference */
    for _, ref := range references {
        fmt.Fprintf(outF, `<li><span onclick="copy(this)">%s (%d)</span>: <a href="%s" target="blank">Strict</a> / <a href="%s" target="blank">Coarse</a>, <code onclick="copy(this)">%s</code></li>`,
        strings.Join(ref.Authors, ", "),
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

