package algos

import (
	"sort"
	"log"
	"context"
	"sync"
	"time"

	"firebase.google.com/go/v4/db"
	"github.com/roblburris/pool-backend/database"
	"googlemaps.github.io/maps"
)

type safeResMap struct {
	mu *sync.Mutex
	data map[string][]int64
}

type safeUserMap struct {
	mu *sync.Mutex
	data map[int64]database.User
}

type sortEntry struct {
	key int64
	value float64
}

type safeComputeRes struct {
	mu *sync.Mutex
	data []sortEntry
}

type safeFinalPairings struct {
	mu *sync.Mutex
	data map[int64]int64
}


// MatchUsers - matches all of the users together
func MatchUsers(ctx context.Context, ref *db.Ref, mapsClient *maps.Client) {
	// Get the groups/clusters of users
	groups := groupUsers(ctx, ref)
	users := database.GetUsers(ctx, ref)

	// Get the final pairings
	pairings := getPairings(ctx, groups, users, mapsClient)
	
	// Take the pairings and write them to the db
	database.UpdatePair(ctx, ref, pairings)
}

// groupUsers - Finds opted in users and maps them to nearest city and stores groups
// of users in the same location in Redis
// Returns a map[string][]int64 (maps cities to groups of users) that can is used in
// GetPairings
func groupUsers(ctx context.Context, ref *db.Ref) map[string][]int64 {
	// Get a list of users from the database and create a slice of keys
	users := database.GetUsers(ctx, ref)
	keys := make([]int64, len(users))

	i := 0
	for k := range users {
		keys[i] = k
		i++
	}

	// Parallely compute city groupings for users 
	safeUsers := safeUserMap{data: users}
	safeRes := safeResMap{data: make(map[string][]int64)}
	parallelReduceToGroups(keys, safeRes, safeUsers)

	return safeRes.data
}

// Parallely computes groups of localized users
func parallelReduceToGroups(uids []int64, res safeResMap, users safeUserMap) {
	// We define the line between creating a new thread to be when len(uids) < 1000
	if len(uids) < 1000 {
		res.mu.Lock()
		for i := 0; i < len(uids); i++ {
			res.data[users.data[uids[i]].TargetCity] = append(res.data[users.data[uids[i]].TargetCity], uids[i])
		}
		res.mu.Unlock()
		return
	}
	
	// If there is still more to process, create new Go routines that call parallelReduce on slices of uids
	go parallelReduceToGroups(uids[0:len(uids) / 2], res, users)
	go parallelReduceToGroups(uids[len(uids) / 2:len(uids)], res, users)
}

// getPairings - For each city (group of clustered users), iterate through and match
// users with one another
// :param cities: a map[string][]int64
// Returns
func getPairings(ctx context.Context, groupings map[string][]int64, userInfo map[int64]database.User, mapsClient *maps.Client) map[int64]int64 {
	var waitGroup sync.WaitGroup
	finPairings := safeFinalPairings{data: make(map[int64]int64)}
	for _, v := range groupings {
		waitGroup.Add(1)
		go getPairingsForCity(ctx, &waitGroup, v, userInfo, mapsClient, finPairings)
	}
	waitGroup.Wait()

	return finPairings.data
}

func getPairingsForCity(ctx context.Context, wg *sync.WaitGroup, cityUsers []int64, userInfo map[int64]database.User, mapsClient *maps.Client, finPairings safeFinalPairings) {
	// Not the best algorithm but we greedily choose best possible match for each
	// person :(
	defer wg.Done()
	for len(cityUsers) > 1 {
		curUser := userInfo[cityUsers[0]]
		safeCurTravelTimes := safeComputeRes{data: make([]sortEntry, 4)}

		var waitGroup sync.WaitGroup
		for i := 1; i < len(cityUsers); i++ {
			waitGroup.Add(1)
			go computeTravelTime(ctx, &waitGroup, curUser.ID, cityUsers[i], userInfo, mapsClient, safeCurTravelTimes)
		}
		waitGroup.Wait()

		// Sort and then match original user to first user in slice
		safeCurTravelTimes.mu.Lock()
		sort.SliceStable(safeCurTravelTimes.data, func(i, j int) bool {
			return safeCurTravelTimes.data[i].value < safeCurTravelTimes.data[j].value
		})

		// We have a match! Add to finPairings and then remove from cityUsers 
		finPairings.mu.Lock()
		finPairings.data[curUser.ID] = safeCurTravelTimes.data[0].key
		finPairings.mu.Unlock()
		cityUsers = cityUsers[1:]
		secondIndex := getIndex(cityUsers, safeCurTravelTimes.data[0].key)
		cityUsers = append(cityUsers[:secondIndex], cityUsers[secondIndex+1:]...)
		safeCurTravelTimes.mu.Unlock()
	}
}

func getIndex(slice []int64, target int64) int {
	for i := 0; i < len(slice); i++ {
		if slice[i] == target {
			return i
		}
	}
	return -1
}

func computeTravelTime(ctx context.Context, waitGroup *sync.WaitGroup, firstUID int64, secondUID int64, userInfo map[int64]database.User, mapsClient *maps.Client, res safeComputeRes) {
	defer waitGroup.Done()
	r := &maps.DistanceMatrixRequest{
		Units: maps.UnitsImperial,
		Origins: []string{userInfo[firstUID].Destination},
		Destinations: []string{userInfo[secondUID].Destination},
	}

	computeRes, err := mapsClient.DistanceMatrix(ctx, r)
	if err != nil {
		log.Println(err)
	}

	res.mu.Lock()
	res.data = append(res.data, sortEntry{key: secondUID, value: time.Duration.Minutes(computeRes.Rows[0].Elements[0].Duration)}) 
	res.mu.Unlock()
}
