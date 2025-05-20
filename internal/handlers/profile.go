package handlers

import (
	"net/http"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/gin-gonic/gin"
)

// ProfileHandler handles profiling-related HTTP requests
type ProfileHandler struct{}

// NewProfileHandler creates a new ProfileHandler instance
func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{}
}

// StartCPUProfile starts CPU profiling for the specified duration
func (h *ProfileHandler) StartCPUProfile(c *gin.Context) {
	// Parse duration from query parameter, default to 30 seconds
	duration := 30 * time.Second
	if d := c.Query("duration"); d != "" {
		if parsedDuration, err := time.ParseDuration(d); err == nil {
			duration = parsedDuration
		}
	}

	// Start CPU profiling
	if err := pprof.StartCPUProfile(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start CPU profile: " + err.Error(),
		})
		return
	}

	// Stop profiling after duration
	time.Sleep(duration)
	pprof.StopCPUProfile()
}

// GetMemoryProfile returns the current memory profile
func (h *ProfileHandler) GetMemoryProfile(c *gin.Context) {
	// Run garbage collection to get accurate memory statistics
	runtime.GC()

	// Write memory profile
	if err := pprof.WriteHeapProfile(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to write memory profile: " + err.Error(),
		})
		return
	}
}

// GetGoroutineProfile returns the current goroutine profile
func (h *ProfileHandler) GetGoroutineProfile(c *gin.Context) {
	// Get goroutine profile
	p := pprof.Lookup("goroutine")
	if p == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get goroutine profile",
		})
		return
	}

	if err := p.WriteTo(c.Writer, 1); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to write goroutine profile: " + err.Error(),
		})
		return
	}
}
