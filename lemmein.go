package main

import (
	"os"
	"log"
	"strings"
	"net/http"
	"text/template"
)


var keymap = map[string][]byte{

	"randy": []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC82w0+Q2j2hfprW2k64tRWgjw9euJiOPJw8JAD6dMj9HeLCrGtjlr+eEi51dSi7/BvGjT0LH1LvNAIgU/I/Bbn99TafcDqo0PZHQ3QqsGh4G8r7O7apcRKxmmHh2bAnMQ3lBvSqnBu5uQ0OBNpvRtmRQLdlvAkh1DzbYSXEy8ypGSmWHkb7X03pEWdIRRL76+Kg2eaYPemzxKf0QcmhocFXjYrR5w3OE12W4DTsFTYXWYTSUZVYMXC62pRDvq10zRXRzGidnW1+d9iFDHqob2yOJ0GGjxW2BV/Nw2GFJncuJUofz7ATGEmIM66B08A8Jchz6Xv00/SNSwmYgfZhxvL randy@beejay"),
	
	"rschwager": []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDOsJ8aR0j69Q49fPS4TTm45xyByAdb5I1JI5p1K1WcAaqMn2ly3YNqa0I8ub4ySuL0U3tpXRsBPjV3S6YbQ0W1BuzebFVOVJtKEk4KOg2VNB1gy9kfrsJDjLlbS97qQnvSZ5IpbqVz61G284u93hRvKhAsAapgQK05RqMXy8YckwvBHjR5OBeGb12y9jL9PQod0Mfdb0MK4HthAzqpZJjgTT8Myh7UPc4BXJf5183LtfK0pqTdvELRk8sYpwEcnmPBXrj7rwUd0Wz3lYcmM8ZKG5XsHavdkER7gr2tuIW9I6xwBibgXsqSkuF8nDAyVIXcvpJ3kxw3rPQZnhgmg/Jb randallschwager@hsph-C1ML5AQXDTY4"),

}

// TODO: dynamically generate shell script url parameter via os.Getenv

var shelltemplate, shelltemplate_err = template.New("shell").Parse(`#!/bin/bash
if ! test -d ~/.ssh; then 
    mkdir -v ~/.ssh; 
    chmod -v 700 ~/.ssh; 
fi
if ! grep -q '{{.Sshkeyfrag}}' ~/.ssh/authorized_keys 2>/dev/null; then 
    echo "{{.Sshkey}}" >> ~/.ssh/authorized_keys;
else
    echo "{{.Querykey}} already has access here"
fi

url="http://{{.Host}}/{{.Querykey}}"
data="user=$USER"

which curl > /dev/null
if test $? -eq 0 ; then 
    curl "$url" -d "$data" 
else 
    wget -O- "$url" --post-data="$data"
fi
`)

type ShellVariables struct {
	Sshkey      string
	Sshkeyfrag  string
	Querykey    string
	Host        string
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("LEMMEIN_HOST")
	if host == "" {
		host = "localhost"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fields := strings.Split(r.URL.String(), "/")
		if len(fields) != 2 {
			log.Print("Received invalid request: ", r)
			http.NotFound(w, r)
			return
		}

		key := fields[1]
		sshkey, present := keymap[key]
		if !present {
			log.Print("Unrecognized user: ", r)
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case "GET":
			vars :=	ShellVariables{ 
				Sshkey:     string(sshkey), 
				Sshkeyfrag: string(sshkey[:50]), 
				Querykey:   key,
				Host:       host,
			}
			shelltemplate.Execute(w, vars)
		case "POST":
			r.ParseForm()
			log.Print("Got data: ", r.Form)
		} 

		return
	})

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
