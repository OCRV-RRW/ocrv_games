package develop_cors

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"net/url"
	"strconv"
	"strings"
)

// ConfigDefault is the default config
var ConfigDefault = cors.Config{
	Next:             nil,
	AllowOriginsFunc: nil,
	AllowOrigins:     "*",
	AllowMethods: strings.Join([]string{
		fiber.MethodGet,
		fiber.MethodPost,
		fiber.MethodHead,
		fiber.MethodPut,
		fiber.MethodDelete,
		fiber.MethodPatch,
	}, ","),
	AllowHeaders:     "Accept-Encoding,Connection,Access-Control-Allow-Origin,Authorization",
	AllowCredentials: true,
	ExposeHeaders:    "",
	MaxAge:           0,
}

// New creates a new middleware handler
func New(config ...cors.Config) fiber.Handler {
	// Set default config
	cfg := ConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]

		// Set default values
		if cfg.AllowMethods == "" {
			cfg.AllowMethods = ConfigDefault.AllowMethods
		}
		// When none of the AllowOrigins or AllowOriginsFunc config was defined, set the default AllowOrigins value with "*"
		if cfg.AllowOrigins == "" && cfg.AllowOriginsFunc == nil {
			cfg.AllowOrigins = ConfigDefault.AllowOrigins
		}
	}

	// Warning logs if both AllowOrigins and AllowOriginsFunc are set
	if cfg.AllowOrigins != "" && cfg.AllowOriginsFunc != nil {
		log.Warn("[CORS] Both 'AllowOrigins' and 'AllowOriginsFunc' have been defined.")
	}

	// allowOrigins is a slice of strings that contains the allowed origins
	// defined in the 'AllowOrigins' configuration.
	allowOrigins := []string{}
	allowSOrigins := []subdomain{}
	allowAllOrigins := false

	// Validate and normalize static AllowOrigins
	if cfg.AllowOrigins != "" && cfg.AllowOrigins != "*" {
		origins := strings.Split(cfg.AllowOrigins, ",")
		for _, origin := range origins {
			if i := strings.Index(origin, "://*."); i != -1 {
				trimmedOrigin := strings.TrimSpace(origin[:i+3] + origin[i+4:])
				isValid, normalizedOrigin := normalizeOrigin(trimmedOrigin)
				if !isValid {
					panic("[CORS] Invalid origin format in configuration: " + trimmedOrigin)
				}
				sd := subdomain{prefix: normalizedOrigin[:i+3], suffix: normalizedOrigin[i+3:]}
				allowSOrigins = append(allowSOrigins, sd)
			} else {
				trimmedOrigin := strings.TrimSpace(origin)
				isValid, normalizedOrigin := normalizeOrigin(trimmedOrigin)
				if !isValid {
					panic("[CORS] Invalid origin format in configuration: " + trimmedOrigin)
				}
				allowOrigins = append(allowOrigins, normalizedOrigin)
			}
		}
	} else if cfg.AllowOrigins == "*" {
		allowAllOrigins = true
	}

	// Strip white spaces
	allowMethods := strings.ReplaceAll(cfg.AllowMethods, " ", "")
	allowHeaders := strings.ReplaceAll(cfg.AllowHeaders, " ", "")
	exposeHeaders := strings.ReplaceAll(cfg.ExposeHeaders, " ", "")

	// Convert int to string
	maxAge := strconv.Itoa(cfg.MaxAge)

	// Return new handler
	return func(c *fiber.Ctx) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Get originHeader header
		originHeader := strings.ToLower(c.Get(fiber.HeaderOrigin))

		// If the request does not have Origin header, the request is outside the scope of CORS
		if originHeader == "" {
			// See https://fetch.spec.whatwg.org/#cors-protocol-and-http-caches
			// Unless all origins are allowed, we include the Vary header to cache the response correctly
			if !allowAllOrigins {
				c.Vary(fiber.HeaderOrigin)
			}

			return c.Next()
		}

		// If it's a preflight request and doesn't have Access-Control-Request-Method header, it's outside the scope of CORS
		if c.Method() == fiber.MethodOptions && c.Get(fiber.HeaderAccessControlRequestMethod) == "" {
			// Response to OPTIONS request should not be cached but,
			// some caching can be configured to cache such responses.
			// To Avoid poisoning the cache, we include the Vary header
			// for non-CORS OPTIONS requests:
			c.Vary(fiber.HeaderOrigin)
			return c.Next()
		}

		// Set default allowOrigin to empty string
		allowOrigin := ""

		// Check allowed origins
		if allowAllOrigins {
			allowOrigin = "*"
		} else {
			// Check if the origin is in the list of allowed origins
			for _, origin := range allowOrigins {
				if origin == originHeader {
					allowOrigin = originHeader
					break
				}
			}

			// Check if the origin is in the list of allowed subdomains
			if allowOrigin == "" {
				for _, sOrigin := range allowSOrigins {
					if sOrigin.match(originHeader) {
						allowOrigin = originHeader
						break
					}
				}
			}
		}

		// Run AllowOriginsFunc if the logic for
		// handling the value in 'AllowOrigins' does
		// not result in allowOrigin being set.
		if allowOrigin == "" && cfg.AllowOriginsFunc != nil && cfg.AllowOriginsFunc(originHeader) {
			allowOrigin = originHeader
		}

		// Simple request
		// Ommit allowMethods and allowHeaders, only used for pre-flight requests
		if c.Method() != fiber.MethodOptions {
			if !allowAllOrigins {
				// See https://fetch.spec.whatwg.org/#cors-protocol-and-http-caches
				c.Vary(fiber.HeaderOrigin)
			}
			setCORSHeaders(c, allowOrigin, "", "", exposeHeaders, maxAge, cfg)
			return c.Next()
		}

		// Pre-flight request

		// Response to OPTIONS request should not be cached but,
		// some caching can be configured to cache such responses.
		// To Avoid poisoning the cache, we include the Vary header
		// of preflight responses:
		c.Vary(fiber.HeaderAccessControlRequestMethod)
		c.Vary(fiber.HeaderAccessControlRequestHeaders)
		c.Vary(fiber.HeaderOrigin)

		setCORSHeaders(c, allowOrigin, allowMethods, allowHeaders, exposeHeaders, maxAge, cfg)

		// Send 204 No Content
		return c.SendStatus(fiber.StatusNoContent)
	}
}

// Function to set CORS headers
func setCORSHeaders(c *fiber.Ctx, allowOrigin, allowMethods, allowHeaders, exposeHeaders, maxAge string, cfg cors.Config) {
	if cfg.AllowCredentials {
		if allowOrigin != "" {
			c.Set(fiber.HeaderAccessControlAllowOrigin, allowOrigin)
			c.Set(fiber.HeaderAccessControlAllowCredentials, "true")
		}
	} else if allowOrigin != "" {
		// For non-credential requests, it's safe to set to '*' or specific origins
		c.Set(fiber.HeaderAccessControlAllowOrigin, allowOrigin)
	}

	// Set Allow-Methods if not empty
	if allowMethods != "" {
		c.Set(fiber.HeaderAccessControlAllowMethods, allowMethods)
	}

	// Set Allow-Headers if not empty
	if allowHeaders != "" {
		c.Set(fiber.HeaderAccessControlAllowHeaders, allowHeaders)
	} else {
		h := c.Get(fiber.HeaderAccessControlRequestHeaders)
		if h != "" {
			c.Set(fiber.HeaderAccessControlAllowHeaders, h)
		}
	}

	// Set MaxAge if set
	if cfg.MaxAge > 0 {
		c.Set(fiber.HeaderAccessControlMaxAge, maxAge)
	} else if cfg.MaxAge < 0 {
		c.Set(fiber.HeaderAccessControlMaxAge, "0")
	}

	// Set Expose-Headers if not empty
	if exposeHeaders != "" {
		c.Set(fiber.HeaderAccessControlExposeHeaders, exposeHeaders)
	}
}

// matchScheme compares the scheme of the domain and pattern
func matchScheme(domain, pattern string) bool {
	didx := strings.Index(domain, ":")
	pidx := strings.Index(pattern, ":")
	return didx != -1 && pidx != -1 && domain[:didx] == pattern[:pidx]
}

// normalizeDomain removes the scheme and port from the input domain
func normalizeDomain(input string) string {
	// Remove scheme
	input = strings.TrimPrefix(strings.TrimPrefix(input, "http://"), "https://")

	// Find and remove port, if present
	if len(input) > 0 && input[0] != '[' {
		if portIndex := strings.Index(input, ":"); portIndex != -1 {
			input = input[:portIndex]
		}
	}

	return input
}

// normalizeOrigin checks if the provided origin is in a correct format
// and normalizes it by removing any path or trailing slash.
// It returns a boolean indicating whether the origin is valid
// and the normalized origin.
func normalizeOrigin(origin string) (bool, string) {
	parsedOrigin, err := url.Parse(origin)
	if err != nil {
		return false, ""
	}

	// Validate the scheme is either http or https
	if parsedOrigin.Scheme != "http" && parsedOrigin.Scheme != "https" {
		return false, ""
	}

	// Don't allow a wildcard with a protocol
	// wildcards cannot be used within any other value. For example, the following header is not valid:
	// Access-Control-Allow-Origin: https://*
	if strings.Contains(parsedOrigin.Host, "*") {
		return false, ""
	}

	// Validate there is a host present. The presence of a path, query, or fragment components
	// is checked, but a trailing "/" (indicative of the root) is allowed for the path and will be normalized
	if parsedOrigin.Host == "" || (parsedOrigin.Path != "" && parsedOrigin.Path != "/") || parsedOrigin.RawQuery != "" || parsedOrigin.Fragment != "" {
		return false, ""
	}

	// Normalize the origin by constructing it from the scheme and host.
	// The path or trailing slash is not included in the normalized origin.
	return true, strings.ToLower(parsedOrigin.Scheme) + "://" + strings.ToLower(parsedOrigin.Host)
}

type subdomain struct {
	// The wildcard pattern
	prefix string
	suffix string
}

func (s subdomain) match(o string) bool {
	return len(o) >= len(s.prefix)+len(s.suffix) && strings.HasPrefix(o, s.prefix) && strings.HasSuffix(o, s.suffix)
}
