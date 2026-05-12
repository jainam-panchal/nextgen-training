package main

import (
	"fmt"
	"time"

	"ride-sharing/internal/config"
	"ride-sharing/internal/dispatch"
	"ride-sharing/internal/driver"
	"ride-sharing/internal/models"
	"ride-sharing/internal/ridequeue"
	"ride-sharing/internal/rider"
	"ride-sharing/internal/rides"
)

func main() {
	cfg := config.DefaultConfig()

	driverStore := driver.NewMemoryStore(cfg)
	riderStore := rider.NewMemoryStore()
	requestQueue := ridequeue.NewRideRequestPriorityQueue()
	activeRideTracker := rides.NewMemoryActiveRideTracker()
	rideHistoryStore := rides.NewMemoryRideHistoryStore()

	dispatcher := dispatch.NewDispatcher(
		cfg,
		driverStore,
		riderStore,
		requestQueue,
		activeRideTracker,
		rideHistoryStore,
	)

	newDriver := &models.Driver{
		ID:       models.NewDriverID(),
		Name:     "Jainam",
		Location: models.Location{Lat: 19.0760, Lng: 72.8777},
		Status:   models.DriverOffline,
		Rating:   4.8,
	}

	if err := driverStore.Register(newDriver); err != nil {
		fmt.Println("register driver error:", err)
		return
	}

	if err := driverStore.GoOnline(newDriver.ID); err != nil {
		fmt.Println("go online error:", err)
		return
	}

	newRider := &models.Rider{
		ID:            models.NewRiderID(),
		Name:          "Jainam",
		Location:      models.Location{Lat: 19.0800, Lng: 72.8800},
		PaymentMethod: "UPI",
	}

	if err := riderStore.Register(newRider); err != nil {
		fmt.Println("register rider error:", err)
		return
	}

	request, err := dispatcher.RequestRide(
		newRider.ID,
		newRider.Location,
		models.Location{Lat: 19.0900, Lng: 72.8900},
		time.Now(),
	)
	if err != nil {
		fmt.Println("request ride error:", err)
		return
	}

	fmt.Println("request created:", request.ID)

	ride, err := dispatcher.ProcessNextRequest(time.Now())
	if err != nil {
		fmt.Println("dispatch error:", err)
		return
	}

	fmt.Println("ride assigned:", ride.ID)
	fmt.Println("driver assigned:", ride.DriverID)
	fmt.Printf("fare: ₹%.2f\n", ride.Fare)

	completedRide, err := dispatcher.CompleteRide(ride.ID, time.Now())
	if err != nil {
		fmt.Println("complete ride error:", err)
		return
	}

	fmt.Println("ride completed:", completedRide.ID)
	fmt.Println("status:", completedRide.Status)

	todayEarnings, err := dispatcher.DriverEarningsToday(newDriver.ID, time.Now())
	if err != nil {
		fmt.Println("earnings error:", err)
		return
	}

	fmt.Printf("driver earnings today: ₹%.2f\n", todayEarnings)
}
