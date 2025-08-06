package main

import (
	"errors"
	"fmt"
)

// Error types for different categories of failures
var (
	ErrFlickrAuth   = errors.New("flickr authentication error")
	ErrFlickrServer = errors.New("flickr server error")
	ErrFlickrUsage  = errors.New("flickr usage error")
	ErrFlickrAPI    = errors.New("flickr api error")
	ErrFileIO       = errors.New("file io error")
	ErrInputs       = errors.New("input validation error")
	ErrUsage        = errors.New("usage error")
)

// Wrap functions for creating errors with context
func WrapFlickrAuth(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, fmt.Errorf("%w: %s", ErrFlickrAuth, err.Error()))
}

func WrapFlickrServer(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, fmt.Errorf("%w: %s", ErrFlickrServer, err.Error()))
}

func WrapFlickrUsage(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, fmt.Errorf("%w: %s", ErrFlickrUsage, err.Error()))
}

func WrapFlickrAPI(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, fmt.Errorf("%w: %s", ErrFlickrAPI, err.Error()))
}

func WrapFileIO(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, fmt.Errorf("%w: %s", ErrFileIO, err.Error()))
}

func WrapInputs(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, fmt.Errorf("%w: %s", ErrInputs, err.Error()))
}

func WrapUsage(msg string) error {
	return fmt.Errorf("%w: %s", ErrUsage, msg)
}

// NewFlickrAuth creates a new Flickr authentication error
func NewFlickrAuth(msg string) error {
	return fmt.Errorf("%w: %s", ErrFlickrAuth, msg)
}

// NewFlickrServer creates a new Flickr server error
func NewFlickrServer(msg string) error {
	return fmt.Errorf("%w: %s", ErrFlickrServer, msg)
}

// NewFlickrUsage creates a new Flickr usage error
func NewFlickrUsage(msg string) error {
	return fmt.Errorf("%w: %s", ErrFlickrUsage, msg)
}

// NewFlickrAPI creates a new Flickr API error
func NewFlickrAPI(msg string) error {
	return fmt.Errorf("%w: %s", ErrFlickrAPI, msg)
}

// NewFileIO creates a new file IO error
func NewFileIO(msg string) error {
	return fmt.Errorf("%w: %s", ErrFileIO, msg)
}

// NewInputs creates a new input validation error
func NewInputs(msg string) error {
	return fmt.Errorf("%w: %s", ErrInputs, msg)
}

// NewUsage creates a new usage error
func NewUsage(msg string) error {
	return fmt.Errorf("%w: %s", ErrUsage, msg)
}

// ClassifyFlickrError classifies a Flickr API error by HTTP status code or error code
func ClassifyFlickrError(statusCode int, errorCode int, message string) error {
	switch statusCode {
	case 401, 403:
		return NewFlickrAuth(message)
	case 400:
		return NewFlickrUsage(message)
	case 500, 501, 502, 503, 504, 505:
		return NewFlickrServer(message)
	default:
		if errorCode >= 400 && errorCode < 500 {
			if errorCode == 401 || errorCode == 403 {
				return NewFlickrAuth(message)
			}
			if errorCode == 400 {
				return NewFlickrUsage(message)
			}
		}
		return NewFlickrServer(message)
	}
}
