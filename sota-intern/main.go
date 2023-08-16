package main

import (
	"encoding/json"
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type HalResponse struct {
    Response Response `json:"response"`
}

type Response struct {
    NumFound int         `json:"numFound"`
    Start int            `json:"start"`
    MaxScore float32     `json:"maxScore"`
    NumFoundExact bool   `json:"numFoundExact"`
    Documents []Document `json:"docs"`
}

type Document struct {
    Title []string    `json:"title_s"`
    Abstract []string `json:"abstract_s"`
    HalId string      `json:"halId_s"`
	Domains []string  `json:"domain_s"`
    SubDate string    `json:"submittedDate_tdate"`
}

func main() {
	// my_keywords := []string{ "reproductibilit%C3%A9", "reproducibility", "control%20theory", "nix", "guix", "cigri", "bag-of-tasks", "simgrid" }
	// my_keywords := []string{ "reproductibilit%C3%A9", "reproducibility", "nix", "guix", "docker" }
	my_keywords := []string{"nixos"}
	domain := "1.info.info-dc"
	// date := "%5BNOW-5YEAR/DAY%20TO%20NOW/HOUR%5D"

	fields := make([]string, 0)
	for _, kw := range my_keywords {
		fields = append(fields, fmt.Sprintf("\"%s\"~", kw))	
	}
	title_request := strings.Join(fields, "||") 
	url := fmt.Sprintf("https://api.archives-ouvertes.fr/search/?q=abstract_t:(%s)&fq=title_t:(%s)&fq=openAccess_bool:true&wt=json&fq=domain_s:%s&fl=title_s,submittedDate_tdate,abstract_s,halId_s,domain_s&rows=100000", title_request, title_request, domain)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("get")
	    panic(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read")
	    panic(err)
	}

  	err = ioutil.WriteFile("output.txt", data, 0644)
    if err != nil {
        panic(err)
    }

	var person HalResponse
	err = json.Unmarshal(data, &person)
	if err != nil {
		fmt.Println("json")
		panic(err)
	}


	for _, doc := range person.Response.Documents {
		fmt.Printf("%-200s [https://hal.science/%s]\n", doc.Title[0], doc.HalId)

		file, err := os.Create(fmt.Sprintf("%s.pdf", doc.HalId))
		if err != nil {
			panic(err)
		}
		resp, err := http.Get(fmt.Sprintf("https://hal.science/%s/document", doc.HalId))
		if err != nil {
			panic(err)
		}
		size, err := io.Copy(file, resp.Body)
		if err != nil {
			panic(err)
		}

		fmt.Println(size)
		
		file.Close()
		resp.Body.Close()
	}

}
