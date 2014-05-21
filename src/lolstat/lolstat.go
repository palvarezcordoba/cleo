package main

import "fmt"
import "query"
import "time"
import "proto"
//import "sort"
import gproto "code.google.com/p/goprotobuf/proto"

type LivePCGL struct {
	Champions	map[proto.ChampionType]LivePCGLRecord
	All			[]uint32
}

type LivePCGLRecord struct {
	Winning 	[]uint32
	Losing		[]uint32
}

// Reads in a ChampionGameList file that can be used for searching.
// TODO: Retrieve the file.
func read_cgl(filename string) LivePCGL {
	pcgl := proto.PackedChampionGameList{}
	live_pcgl := LivePCGL{}

//	gproto.Unmarshal(bytes, &cgl)	
	live_pcgl.All = pcgl.All

	for _, record := range pcgl.Champions {
		live_pcgl.Champions[*record.Champion] = LivePCGLRecord{record.Winning , record.Losing}
//		live_pcgl.Champions[*record.Champion].Winning = *record.Winning
//		live_pcgl.Champions[*record.Champion].Losing = *record.Losing
	}
	
	return live_pcgl
}

func main() {
	// Inputs
	query_requests := make(chan query.GameQueryRequest, 100)
	// Outputs
	query_completions := make(chan query.GameQueryResponse, 100)

	fmt.Printf("Loading gamelog.\n")
	cgl := read_cgl("latest.cgl")

	qm := query.QueryManager{}
	qm.Connect()

	// Kick off some goroutines that can handle queries.
	for i := 0; i < 1; i++ {
		go query_handler(query_requests, &cgl, query_completions)
	}

	// Kick off one goroutine that can handle responding to queries.
	go query_responder(query_completions)

	// Infinitely loop through queries as they come in. Currently this
	// will only handle one at a time but should be trivial to parallelize
	// once the time is right.
	for {
		gqr := query.GameQueryRequest{}
		gqr.Query = &proto.GameQuery{}
		gqr.Id = 1
		fmt.Println(gqr.Query)
		gqr.Query.Winners = append(gqr.Query.Winners, proto.ChampionType_THRESH)

		query_requests <- gqr
//		qm.Await()
//		query_requests <- qm.Await()
		time.Sleep(10 * time.Second)
	}
}

// A query handler filters down the list of game ID's to the set identified in
// the provided query. They are run as goroutines and can handle a single
// query at a time. They each make a copy of the lists in the CGl so that
// all queries are independent and unaffected by others.
//
// Two values need to be computed: the MATCHING games and the ELIGIBLE games.
//   - Matching games are those that have all of the requested players
//     on the requested teams (numerator).
//   - Eligible games are those that have all of the requested players
//     on one team or another (denominator).
//
// The general algorithm for computing each is as follows:
//
// MATCHING
// Start with a list of all games. For each winning champion, find the
// overlap between the full set with the winning game set for that champion.
// Then do the same thing for all losing champions.
//
// ELIGIBLE
// Start with a list of all games. For each winning champion, find the
// overlap between the full set and the losing game set for that champion.
// Then do the same thing for all losing champions (find the winning set
// for them). Then merge the output from the MATCHING set with the lists
// from the ELIGIBLE set to produce the final ELIGIBLE set.  
func query_handler(input chan query.GameQueryRequest, pcgl *LivePCGL, output chan query.GameQueryResponse) {
	for {
		request := <-input
		fmt.Println("Handling query #", request.Id)
		
		// Eligible gamelist contains all games that match, irrespective of team.
		eligible_gamelist := pcgl.All
		
		// Matching gamelist contains all games that match, respective of team.
		matching_gamelist := pcgl.All

		// Merge all game ID's, first matching the winning parameters. 
		for _, champion := range request.Query.Winners {
			// Update the matching gamelist to include just the overlap between these two lists.
			overlap(&matching_gamelist, pcgl.Champions[champion].Winning)
			overlap(&eligible_gamelist, pcgl.Champions[champion].Losing)
		}
		
		// Then match all losers.
		for _, champion := range request.Query.Losers {
			overlap(&matching_gamelist, pcgl.Champions[champion].Losing)
			overlap(&eligible_gamelist, pcgl.Champions[champion].Winning)
		}
		
		//// Step #2: Eligible set (not incl. team) ////
		eligible_gamelist = merge(eligible_gamelist, matching_gamelist)
		//eligible_gamelist = append(eligible_gamelist, matching_gamelist...)
		
		// Prepare the response.
		response := query.GameQueryResponse{}
		response.Id = request.Id
		response.Conn = request.Conn
		
		response.Response.Available = gproto.Uint32( uint32(len(eligible_gamelist)) )
		response.Response.Matching = gproto.Uint32( uint32(len(matching_gamelist)) )
		response.Response.Total = gproto.Uint32( uint32(len(pcgl.All)) )
		
		output <- response
	}
}

func query_responder(input chan query.GameQueryResponse) {
	for {
		response := <-input

		fmt.Println(fmt.Sprintf("Responding to query #%d", response.Id))
		// TODO: Send the response.
//		response.Conn.Send(&gproto.Marshal(response.Response))
	}
}

// Overlap accepts two lists of uints and reduces FIRST to the overlap
// between both lists. 
// Assumes that both lists are ordered.
func overlap(first *[]uint32, second []uint32) {
	// parallel_counter indexes into SECOND and may move at a different
	// rate than i.
	parallel_counter := 0
	
	for i := 0; i < len(*first); i++ {
		// Loop through until the second array's value is greater than or
		// equal to the primary array. We should not reset this counter
		// variable.
		for second[parallel_counter] < (*first)[i] {
			if parallel_counter + 1 < len(second) {
				parallel_counter += 1
			} else {
			// If parallel_counter is as big as it can get then none of
			// the other numbers in FIRST can overlap.
				(*first) = (*first)[:i]
				return
			}			
		}
		// Once the secondary index catches up, if it's beyond the primary
		// then the primary doesn't exist. If they're equivalent then we
		// keep the primary value.
		if second[parallel_counter] > (*first)[i] {			
			(*first)[i] = 0
			(*first) = append( (*first)[:i], (*first)[i+1:]... )
			
			fmt.Println(*first)
			i -= 1
		}
	}
}

func merge(first []uint32, second []uint32) []uint32 {
	full := make([]uint32, 0, len(first) + len(second))
	
	first_i := 0
	second_i := 0
	
	// Move through the list until we get to the end of one of them.
	for first_i < len(first) && second_i < len(second) {
		fmt.Println(fmt.Sprintf("first=%d, second=%d", first_i, second_i))
		// If next value in FIRST is less than next value in SECOND,
		// copy value from FIRST and move on.
		if first[first_i] < second[second_i] {
			full = append(full, first[first_i])
			
			first_i += 1
		// If the two values are the same, copy one of them over. This
		// will remove duplicates. 
		} else if first[first_i] == second[second_i] {
			full = append(full, first[first_i])
			
			first_i += 1
			second_i += 1
		// Otherwise if FIRST > SECOND, copy over second.
		} else {
			full = append(full, second[second_i])
			
			second_i += 1
		}
	}
	
	// Copy over all remaining values from FIRST and SECOND.
	if first_i < len(first) {
		full = append(full, first[first_i:]...)
	}
	
	if second_i < len(second) {
		full = append(full, second[second_i:]...)
	}
	
	return full
}
