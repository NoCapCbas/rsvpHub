package main

import (
  "fmt"
  "html/template"
  "net/http"
  "github.com/skip2/go-qrcode"
  "strconv"
)

type Rsvp struct {
  FirstName string 
  LastName string
  GenderGuess string 
  WillAttend bool
  NumberOfPlannedGuest int
}

var responses = make([]*Rsvp, 0, 100)
var templates = make(map[string]*template.Template, 3)

func loadTemplates() {
  // slice of template names
  templateNames := [5]string { "welcome", "form", "thanks", "sorry", "list" }
  // range through template names parsing each with layout.html
  for index, name := range templateNames {
    t, err := template.ParseFiles("src/templates/layout.html", "src/templates/" + name + ".html")
    if (err == nil) {
      templates[name] = t
      fmt.Println("Loaded template", index, name)
    } else {
      panic(err)
    }
  }
}

func welcomeHandler(writer http.ResponseWriter, request *http.Request) {
  templates["welcome"].Execute(writer, nil)
}

func listHandler(writer http.ResponseWriter, request *http.Request) {

  templates["list"].Execute(writer, responses)
}

type formData struct {
  *Rsvp
  Errors []string
}

func formHandler(writer http.ResponseWriter, request *http.Request) {
  if request.Method == http.MethodGet {

    templates["form"].Execute(writer, formData {
      Rsvp: &Rsvp{}, Errors: []string {},
    })

  } else if request.Method == http.MethodPost {
    err := request.ParseForm()
    if err != nil {
      http.Error(writer, fmt.Sprintf("Failed to parse form: %v", err), http.StatusInternalServerError)
    }
    guestCountStr := request.Form["guestcount"][0]
    guestCount, err := strconv.Atoi(guestCountStr)
    if err != nil {
      http.Error(writer, "Invalid guest count value", http.StatusBadRequest)
      return
    }
    responseData := Rsvp {
      FirstName: request.Form["first-name"][0],
      LastName: request.Form["last-name"][0],
      GenderGuess: request.Form["genderguess"][0],
      WillAttend: request.Form["willattend"][0] == "true",
      NumberOfPlannedGuest: guestCount,
    }

    errors := []string {}
    if responseData.FirstName == "" {
      errors = append(errors, "Please enter your first name")
    }
    if responseData.LastName == "" {
      errors = append(errors, "Please enter your last name")
    }
    if len(errors) > 0 {
      templates["form"].Execute(writer, formData {
        Rsvp: &responseData, Errors: errors,
      })
    } else {
      responses = append(responses, &responseData)
      if responseData.WillAttend {
        templates["thanks"].Execute(writer, responseData.FirstName)
      } else {
        templates["sorry"].Execute(writer, responseData.FirstName)
      }
    }
  }
}

func shoppingListQRGenHandler(writer http.ResponseWriter, request *http.Request) {
  if request.Method == http.MethodGet {
    url := "https://www.amazon.com/baby-reg/heather-diaz-december-2024-summerville/1RMR4ST4YLPRN?ref_=cm_sw_r_mwn_dp_NQMGNHAGPDJ4RE18DJX0&language=en_US"
    png, err := qrcode.Encode(url, qrcode.Medium, 256)
    if err != nil {
      http.Error(writer, fmt.Sprintf("Failed to generate QR code: %v", err), http.StatusInternalServerError)
      return
    }

    writer.Header().Set("Content-Type", "image/png")
    writer.Write(png)

  }
}

func main() {
  loadTemplates()

  // Serve static files 
  fs := http.FileServer(http.Dir("src"))
  http.Handle("/src/", http.StripPrefix("/src/", fs))

  http.HandleFunc("/", welcomeHandler)
  http.HandleFunc("/list", listHandler)
  http.HandleFunc("/form", formHandler)
  http.HandleFunc("/qrcode", shoppingListQRGenHandler)

  err := http.ListenAndServe(":8001", nil)
  if (err != nil) {
    fmt.Println(err)
  }
}
