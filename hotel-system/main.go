package main

import "net/http"

func createReservationHandler(w http.ResponseWriter, r *http.Request) {
	// Validate json body

	// validate reservation token

	// validate if room is availabe for date start until end

	// If room is available, create reservation

	// Get latest reservation version

	// update version reservation

	w.Write([]byte("Hello, World!"))
}

func main() {
	http.HandleFunc("/reservation", createReservationHandler)
	http.ListenAndServe(":8080", nil)

	// endpoint to make reservation

	// send, validate reserve-token

	// validate reservation body json

	// validate if room is available

	// create reservation

	// update token user
}
