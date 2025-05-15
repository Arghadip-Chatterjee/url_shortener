package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type ShortenRequest struct {
	URL    string `json:"url"`
	UserID string `json:"user_id"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

type ClickAnalytics struct {
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Country   string    `json:"country"`
	Browser   string    `json:"browser"`
}

type URLData struct {
	OriginalURL string           `json:"original_url"`
	Clicks      []ClickAnalytics `json:"clicks"`
}

var (
	mongoClient    *mongo.Client
	urlCollection  *mongo.Collection
	userCollection *mongo.Collection
)

func generateShortID(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func initMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://arghadipchatterjee2016:Hibye2026@cluster0.9xl9itf.mongodb.net/url_shortener?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}
	mongoClient = client
	urlCollection = client.Database("url_shortener").Collection("urls")
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Use user_id from the request body
	userID := req.UserID
	if userID == "" {
		http.Error(w, "Missing user_id", http.StatusUnauthorized)
		return
	}

	shortID := generateShortID(6)
	urlData := URLData{
		OriginalURL: req.URL,
		Clicks:      []ClickAnalytics{},
	}

	fmt.Println(urlData)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := urlCollection.InsertOne(ctx, bson.M{
		"_id":          shortID,
		"original_url": req.URL,
		"clicks":       []ClickAnalytics{},
		"user_id":      userID,
	})
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	resp := ShortenResponse{ShortURL: r.Host + "/" + shortID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortID := strings.TrimPrefix(r.URL.Path, "/")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var data URLData
	err := urlCollection.FindOne(ctx, bson.M{"_id": shortID}).Decode(&data)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	analytics := ClickAnalytics{
		Timestamp: time.Now(),
		IP:        getRealIP(r),
		UserAgent: r.UserAgent(),
		Country:   getCountryFromIP(r.RemoteAddr),
		Browser:   getBrowserFromUA(r.UserAgent()),
	}
	_, err = urlCollection.UpdateOne(ctx, bson.M{"_id": shortID}, bson.M{"$push": bson.M{"clicks": analytics}})
	http.Redirect(w, r, data.OriginalURL, http.StatusFound)
}

func analyticsHandler(w http.ResponseWriter, r *http.Request) {
	shortID := strings.TrimPrefix(r.URL.Path, "/analytics/")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var data URLData
	err := urlCollection.FindOne(ctx, bson.M{"_id": shortID}).Decode(&data)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.Clicks)
}

func getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain multiple IPs)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// The first IP in the list is the real client IP
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	// Fallback to RemoteAddr (may include port)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return ip
	}
	return r.RemoteAddr
}

func getCountryFromIP(ip string) string {
	url := "https://ip-geolocation-find-ip-location-and-ip-info.p.rapidapi.com/backend/ipinfo/?ip=" + ip
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "Unknown"
	}
	request.Header.Add("x-rapidapi-key", "2c80fc31e8mshf8fc6792897f5b8p146d99jsn5c9fe3bca683")
	request.Header.Add("x-rapidapi-host", "ip-geolocation-find-ip-location-and-ip-info.p.rapidapi.com")
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return "Unknown"
	}
	defer response.Body.Close()
	var result struct {
		CountryName string `json:"country_name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "Unknown"
	}
	if result.CountryName == "" {
		return "Unknown"
	}
	return result.CountryName
}

func getBrowserFromUA(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "chrome"):
		return "Chrome"
	case strings.Contains(ua, "firefox"):
		return "Firefox"
	case strings.Contains(ua, "safari"):
		return "Safari"
	case strings.Contains(ua, "edge"):
		return "Edge"
	case strings.Contains(ua, "opera"):
		return "Opera"
	default:
		return "Other"
	}
}

type FrontendAnalyticsRequest struct {
	ShortID string           `json:"short_id"`
	Events  []ClickAnalytics `json:"events"`
}

func frontendAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	var req FrontendAnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Fill missing IP and country fields for each event
	for i := range req.Events {
		if req.Events[i].IP == "" {
			req.Events[i].IP = getRealIP(r)
		}
		if req.Events[i].Country == "" {
			req.Events[i].Country = getCountryFromIP(req.Events[i].IP)
		}
		if req.Events[i].Timestamp.IsZero() {
			req.Events[i].Timestamp = time.Now()
		}
	}
	_, err := urlCollection.UpdateOne(ctx, bson.M{"_id": req.ShortID}, bson.M{"$push": bson.M{"clicks": bson.M{"$each": req.Events}}})
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getOriginalURLHandler(w http.ResponseWriter, r *http.Request) {
	shortID := strings.TrimPrefix(r.URL.Path, "/api/original/")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var data URLData
	err := urlCollection.FindOne(ctx, bson.M{"_id": shortID}).Decode(&data)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"original_url": data.OriginalURL})
}

func userUrlsHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/api/user-urls/")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := urlCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)
	var urls []bson.M
	if err := cursor.All(ctx, &urls); err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"urls": urls})
}

func InitUserCollection() {
	userCollection = mongoClient.Database("url_shortener").Collection("users")
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var existing User
	err := userCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existing)
	if err == nil {
		http.Error(w, "Email already registered", http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user := User{Name: req.Name, Email: req.Email, Password: string(hash)}
	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Signup successful"})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user User
	err := userCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful", "user_id": user.ID, "name": user.Name, "email": user.Email})
}

func main() {
	initMongo()
	InitUserCollection()
	defer func() {
		if mongoClient != nil {
			_ = mongoClient.Disconnect(context.Background())
		}
	}()

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		shortenHandler(w, r)
	})
	http.HandleFunc("/analytics/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		analyticsHandler(w, r)
	})

	http.HandleFunc("/frontend-analytics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		frontendAnalyticsHandler(w, r)
	})
	http.HandleFunc("/api/original/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		getOriginalURLHandler(w, r)
	})
	// Remove the /api/user/ dynamic route handler
	http.HandleFunc("/api/user-urls/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		userUrlsHandler(w, r)
	})
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		signupHandler(w, r)
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		loginHandler(w, r)
	})
	log.Println("Backend running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type User struct {
	ID       string `bson:"_id,omitempty" json:"id"`
	Name     string `bson:"name" json:"name"`
	Email    string `bson:"email" json:"email"`
	Password string `bson:"password" json:"password"`
}
