package main

import (
  "fmt"
  // "html/template"
  "net/http"
  // "github.com/skip2/go-qrcode"
  "time"
  "sync"
  "strings"
  "encoding/json"
  "encoding/base64"
  // "math/rand"
  "context"
  "errors"
  "crypto/hmac"
  "crypto/sha256"
  "github.com/google/uuid"
  _ "github.com/lib/pq"
  "database/sql"
  "log"
  "golang.org/x/crypto/bcrypt"
)

// MODELS
// User Model
type User struct {
  ID string `json:"-"`
  Email string `json:"email"`
  Password string `json:"password"`
}

var jwtSecretKey = []byte("fjdksaceoaphfionamcieoapfinca")
// Claim Model for jwt
type Claim struct {
  Email string `json:"email"`
  Expires int64 `json:"expires`
}

// Users hosting events relation Model
type HostToEvents struct {
  UserID string `json:"user_id"`
  EventID string `json:"event_id"`
}
var hostToEventsStore = make(map[string]*HostToEvents)

// Event Model
type Event struct {
  ID string `json:"-"`
  Name string  `json:"name"`
  Description string `json:"description"`
}
var eventStore = make(map[string]*Event)

// Rsvp Model
type Rsvp struct {
  ID string `json:"id"`
  FirstName string `json:"first_name"`
  LastName string `json:"last_name"`
  WillAttend bool `json:"will_attend"`
}
// Rsvps Attending Events Relation Model
type RsvpsAttendingEvents struct {
  RsvpID string `json:"rsvp_id"`
  EventID string `json:"event_id"`
}

var responses = make([]*Rsvp, 0, 100)
var mu sync.Mutex
var db *sql.DB
// DATABASE
func connectToDB() {
  // Define your connection string
  connStr := ""

  // Open the database
  var err error
  db, err = sql.Open("postgres", connStr)
  if err != nil {
    log.Fatalf("Error opening database: %v\n", err)
  }

  // Test the connection
  err = db.Ping()
  if err != nil {
    log.Fatalf("Error connecting to database: %v\n", err)
  }

  fmt.Println("Successfully connected to the database!")

}

// HANDLERS

func getEventsHandler(w http.ResponseWriter, r *http.Request) {

  if r.Method != http.MethodGet {
    http.Error(w, "Invalide request method", http.StatusMethodNotAllowed)
  }
  
  email := r.Context().Value("email").(string)

  rows, err := db.Query(`
  SELECT *
  FROM EVENTS 
  `)
  

}

// func readEventHandler(w http.ResponseWriter, r *http.Request) {
//   if r.Method != http.MethodGet {
//     http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
//     return
//   } 
//   
//   email := r.Context().Value("email").(string)
//   hostToEvent := hostToEventsStore[userStore[email].ID]
//   event_id := hostToEvent.EventID
//   event := eventStore[event_id] 
//
//   jsonEvent, err := json.Marshal(event)
//   if err != nil {
//     http.Error(w, "Error marshalling to JSON:", http.StatusInternalServerError)
//     return
//   }
//
//   w.Header().Set("Content-Type", "application/json")
//   w.Write(jsonEvent)
//
// }

// func createEventHandler(w http.ResponseWriter, r *http.Request) {
//
//   if r.Method != http.MethodPost {
//     http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
//     return
//   }
//
//   email := r.Context().Value("email").(string)
//  
//   var event Event
//   err := json.NewDecoder(r.Body).Decode(&event)
//   if err != nil {
//     http.Error(w, "Invalid request body", http.StatusBadRequest)
//     return
//   }
//
//   mu.Lock()
//   defer mu.Unlock()
//
//   if _, exists := eventStore[event.Name]; exists {
//     http.Error(w, "Event name already exists", http.StatusBadRequest)
//     return
//   }
//
//   event.ID = uuid.New().String()
//   
//   eventStore[event.ID] = &event
//   fmt.Printf("New Event Created (%s) id (%s) by %s\n", event.Name, event.ID, email)
//
//   // Host to event relation record 
//   hostToEvent := HostToEvents{
//     UserID: userStore[email].ID, 
//     EventID: event.ID,
//   }
//   hostToEventsStore[userStore[email].ID] = &hostToEvent 
//   fmt.Printf("Host - Event relation created\n")
//   w.WriteHeader(http.StatusCreated)
// }

func registerHandler(w http.ResponseWriter, r *http.Request) {

  if r.Method != http.MethodPost {
    http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    return
  }

  var user User
  err := json.NewDecoder(r.Body).Decode(&user)
  if err != nil {
    http.Error(w, "Invalid request body", http.StatusBadRequest)
    return
  }

  // check if email already exists 
  var exists bool
  err = db.QueryRow(`
  SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);
  `, user.Email).Scan(&exists)

  if exists {
    http.Error(w, "Email already exists", http.StatusBadRequest)
    return
  }

  user.ID = uuid.New().String()
  hashedPassword, err := hashPassword(user.Password)
  if err != nil {
    fmt.Println("Error hashing password:", err)
    return
  }
  user.Password = hashedPassword
  
  // insert user into db
  _, err = db.Exec(`
  INSERT INTO users (id, email, password) VALUES ($1, $2, $3)
  `, user.ID, user.Email, user.Password)
  if err != nil {
    http.Error(w, "Error inserting user", http.StatusInternalServerError)
    return
  }

  fmt.Printf("New User Registered: %s\n", user.Email)
  w.WriteHeader(http.StatusCreated)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodPost {
    http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    return
  }

  var user User
  err := json.NewDecoder(r.Body).Decode(&user)
  if err != nil {
    http.Error(w, "Invalid request body", http.StatusBadRequest)
    return
  }

  mu.Lock()
  defer mu.Unlock()

  var exists bool
  err = db.QueryRow(`
  SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);
  `, user.Email).Scan(&exists)

  if !exists {
    http.Error(w, "Email does not exist. Please register your email.", http.StatusBadRequest)
    return
  }

  var hashedPassword string
  err = db.QueryRow(`
  SELECT password FROM users WHERE email = $1;
  `, user.Email).Scan(&hashedPassword)

  err = checkPassword(user.Password, hashedPassword)
  if err != nil {
    http.Error(w, "Invalid password", http.StatusUnauthorized)
    return
  }

  token, err := generateJWT(user.Email)
  if err != nil {
    http.Error(w, "Failed to generate token", http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
  email := r.Context().Value("email").(string)
  w.Write([]byte(fmt.Sprintf("Hello, %s! You have accessed a protected route.", email)))
}


// func welcomeHandler(writer http.ResponseWriter, request *http.Request) {
//   templates["welcome"].Execute(writer, nil)
// }
//
// func listHandler(writer http.ResponseWriter, request *http.Request) {
//
//   templates["list"].Execute(writer, responses)
// }
//
// type formData struct {
//   *Rsvp
//   Errors []string
// }
//
// func formHandler(writer http.ResponseWriter, request *http.Request) {
//   if request.Method == http.MethodGet {
//
//     templates["form"].Execute(writer, formData {
//       Rsvp: &Rsvp{}, Errors: []string {},
//     })
//
//   } else if request.Method == http.MethodPost {
//     err := request.ParseForm()
//     if err != nil {
//       http.Error(writer, fmt.Sprintf("Failed to parse form: %v", err), http.StatusInternalServerError)
//     }
//     responseData := Rsvp {
//       FirstName: request.Form["first-name"][0],
//       LastName: request.Form["last-name"][0],
//       GenderGuess: request.Form["genderguess"][0],
//       WillAttend: request.Form["willattend"][0] == "true",
//     }
//
//     errors := []string {}
//     if responseData.FirstName == "" {
//       errors = append(errors, "Please enter your first name")
//     }
//     if responseData.LastName == "" {
//       errors = append(errors, "Please enter your last name")
//     }
//     if len(errors) > 0 {
//       templates["form"].Execute(writer, formData {
//         Rsvp: &responseData, Errors: errors,
//       })
//     } else {
//       responses = append(responses, &responseData)
//       if responseData.WillAttend {
//         templates["thanks"].Execute(writer, responseData.FirstName)
//       } else {
//         templates["sorry"].Execute(writer, responseData.FirstName)
//       }
//     }
//   }
// }
//
// func shoppingListQRGenHandler(writer http.ResponseWriter, request *http.Request) {
//   if request.Method == http.MethodGet {
//     url := "https://www.amazon.com/baby-reg/heather-diaz-december-2024-summerville/1RMR4ST4YLPRN?ref_=cm_sw_r_mwn_dp_NQMGNHAGPDJ4RE18DJX0&language=en_US"
//     png, err := qrcode.Encode(url, qrcode.Medium, 256)
//     if err != nil {
//       http.Error(writer, fmt.Sprintf("Failed to generate QR code: %v", err), http.StatusInternalServerError)
//       return
//     }
//
//     writer.Header().Set("Content-Type", "image/png")
//     writer.Write(png)
//
//   }
// }

func main() {
  // connect to remote db
  connectToDB() 

  // close postgres db
  defer db.Close()
  // user registration
  http.HandleFunc("/register", registerHandler)
  http.HandleFunc("/login", loginHandler)
  http.HandleFunc("/protected", authMiddleware(protectedHandler))

  // event crud handlers
  http.HandleFunc("/events", authMiddleware(getEventsHandler))
  // http.HandleFunc("/event/create", authMiddleware(createEventHandler))
  // http.HandleFunc("/event/read", authMiddleware(readEventHandler))
  // http.HandleFunc("/event/update")
  // http.HandleFunc("/event/delete")



  // http.HandleFunc("/", welcomeHandler)
  // http.HandleFunc("/list", listHandler)
  // http.HandleFunc("/form", formHandler)
  // http.HandleFunc("/qrcode", shoppingListQRGenHandler)

  fmt.Println("Server running on port 8001...")
  err := http.ListenAndServe(":8001", nil)
  if (err != nil) {
    fmt.Println(err)
  }
}

// MIDDLEWARE
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
      http.Error(w, "Missing token", http.StatusUnauthorized)
      return
    }

    token := strings.TrimPrefix(authHeader, "Bearer ")
    claim, err := parseJWT(token)
    if err != nil {
      http.Error(w, "Invalid token", http.StatusUnauthorized)
    }

    if time.Now().Unix() > claim.Expires {
      http.Error(w, "Token expired", http.StatusUnauthorized)
      return
    }

    ctx := r.Context()
    ctx = context.WithValue(ctx, "email", claim.Email)
    r = r.WithContext(ctx)

    next(w, r)
  }

}

// UTILS
func generateJWT(email string) (string, error) {
  expires := time.Now().Add(1 * time.Hour).Unix()
  claim := Claim{
    Email: email,
    Expires: expires,
  } 
  
  // Encode claim into JSON
  claimJSON, err := json.Marshal(claim)
  if err != nil {
    return "Error encoding claim", err
  }
 
  // header 
  header := base64Encode([]byte(`{"alg":"none"}`))
  // payload
  payload := base64Encode(claimJSON)
  // Generate random signature
  h := hmac.New(sha256.New, jwtSecretKey)
  h.Write([]byte(payload))
  signature := base64Encode(h.Sum(nil))
  

  // Create the token as "header.payload.signature"
  token := fmt.Sprintf("%s.%s.%s", header, payload, signature)
  return token, nil

}

func parseJWT(token string) (*Claim, error) {
  parts := strings.Split(token, ".")
  if len(parts) != 3 {
    return nil, errors.New("Invalid token format")
  }

  claimJSON, err := base64Decode(parts[1])
  if err != nil {
    return nil, err
  }

  var claim Claim
  err = json.Unmarshal(claimJSON, &claim)
  if err != nil {
    return nil, err
  }

  return &claim, nil
}

func base64Encode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func base64Decode(data string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(data + strings.Repeat("=", (4-len(data)%4)%4))
}

// Function to hash a password using bcrypt
func hashPassword(password string) (string, error) {
	// Generate "cost" factor for bcrypt
	cost := bcrypt.DefaultCost

	// Generate a hashed representation of the password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	// Encode hashed password as a base64 string
	hashedPassword := string(hash)
	return hashedPassword, nil
}

// Function to check if a password matches a hashed password
func checkPassword(password, hashedPassword string) error {
	// Compare the provided password with the hashed password
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return err // Passwords do not match
	}
	return nil // Passwords match
}
