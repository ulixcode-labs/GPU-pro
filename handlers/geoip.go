package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// GeoLocation represents geographical location data
type GeoLocation struct {
	IP          string  `json:"ip"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ISP         string  `json:"isp"`
}

// IPAPIResponse is the response from ip-api.com
type IPAPIResponse struct {
	Query       string  `json:"query"`
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
}

var (
	geoCache   = make(map[string]*GeoLocation)
	geoCacheMu sync.RWMutex
	httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}
)

// IsPrivateIP checks if an IP is private/local
func IsPrivateIP(ip string) bool {
	// Remove port if present
	host := ip
	if strings.Contains(ip, ":") {
		host, _, _ = net.SplitHostPort(ip)
	}

	parsedIP := net.ParseIP(host)
	if parsedIP == nil {
		return true
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fe80::/10",
		"fc00::/7",
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network != nil && network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// ExtractIP extracts IP address from connection address string
func ExtractIP(addr string) string {
	// Handle formats like "192.168.1.1:443" or "[::1]:80"
	if strings.HasPrefix(addr, "[") {
		// IPv6
		endBracket := strings.Index(addr, "]")
		if endBracket > 0 {
			return addr[1:endBracket]
		}
	}

	// IPv4
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// No port, return as-is
		return addr
	}
	return host
}

// LookupGeoLocation performs IP geolocation lookup
func LookupGeoLocation(ip string) (*GeoLocation, error) {
	// Check cache first
	geoCacheMu.RLock()
	if cached, ok := geoCache[ip]; ok {
		geoCacheMu.RUnlock()
		return cached, nil
	}
	geoCacheMu.RUnlock()

	// Use ip-api.com (free, no API key needed, 45 req/min limit)
	url := "http://ip-api.com/json/" + ip + "?fields=status,country,countryCode,region,regionName,city,lat,lon,isp,org,as,query"

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp IPAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if apiResp.Status != "success" {
		return nil, nil
	}

	geoLoc := &GeoLocation{
		IP:          ip,
		Country:     apiResp.Country,
		CountryCode: apiResp.CountryCode,
		Region:      apiResp.RegionName,
		City:        apiResp.City,
		Latitude:    apiResp.Lat,
		Longitude:   apiResp.Lon,
		ISP:         apiResp.ISP,
	}

	// Cache the result
	geoCacheMu.Lock()
	geoCache[ip] = geoLoc
	geoCacheMu.Unlock()

	return geoLoc, nil
}

// GetConnectionGeoLocations gets geolocation for all external IPs in connections
func GetConnectionGeoLocations(connections []NetworkConnection) map[string]*GeoLocation {
	locations := make(map[string]*GeoLocation)
	uniqueIPs := make(map[string]bool)

	// Extract unique external IPs
	for _, conn := range connections {
		foreignIP := ExtractIP(conn.ForeignAddr)
		
		// Skip private/local IPs and already processed IPs
		if !IsPrivateIP(foreignIP) && !uniqueIPs[foreignIP] && foreignIP != "" && foreignIP != "*" {
			uniqueIPs[foreignIP] = true
		}
	}

	// Lookup geolocation for each unique IP (with rate limiting)
	// ip-api.com allows 45 requests per minute for free
	count := 0
	maxPerBatch := 15 // Conservative limit

	for ip := range uniqueIPs {
		if count >= maxPerBatch {
			break // Avoid rate limiting
		}

		geoLoc, err := LookupGeoLocation(ip)
		if err == nil && geoLoc != nil {
			locations[ip] = geoLoc
			count++
		}

		// Small delay to avoid rate limiting
		if count < maxPerBatch {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return locations
}
