package main

import "sort"
import "fmt"


func deepCopyHand (hand []Card) *[]Card {
  newHand := make([]Card, len(hand));
  for i,card := range hand {
    newCard := Card{Suit: card.Suit, Rank:card.Rank}
    newHand[i] = newCard
  }
  return &newHand
}

func isValidRun(cards []Card) bool {
  for i := 1; i < len(cards); i++ {
    if cards[i].Rank == cards[i-1].Rank + 1 && cards[i].Suit == cards[i+1].Suit {

    }

  }
}

func findHandScore(hand []Card) []Card {
	var sets [][]Card
	//first sorting hand by ca
	sort.Slice(hand, func(i, j int) bool {
		return hand[i].Rank < hand[j].Rank
	})
	//now finding sets within those cards if there is three
	curRank := hand[0].Rank
	streak := 0
	for i, card := range hand {
		if card.Rank != curRank {
			if streak >= 3 {
        aSet := hand[i-streak:i]
        aSetCopy := deepCopyHand(aSet)
        sets = append(sets, *aSetCopy)
			}
			streak = 1
			curRank = card.Rank
			continue
		}
		streak++
	}
	if streak >= 3 {
    aSet := hand[len(hand)-streak:]
    aSetCopy := deepCopyHand(aSet)
		sets = append(sets, *aSetCopy)
	}

	//second sorting to find if there is runs
	sort.Slice(hand, func(i, j int) bool {
		if hand[i].Suit == hand[j].Suit {
			return hand[i].Rank < hand[j].Rank
		}
		return hand[i].Suit < hand[j].Suit
	})
  fmt.Println(hand)
  //sliding window approach 
  left := 0
  right := 1
  for right > len(hand) {
    rightCard := hand[right]
    leftCard := hand[left]


  } 


  //for i, card := range hand {
		//if i > 2 && (card.Rank != lastRank + 1 || card.Suit != lastSuit) {
  //    //getting every possible set in this range 
  //    fmt.Println("here", i)
  //    endIndex := i - 1
  //    for windowSize:=3 ; windowSize < streak; windowSize++ {
  //      for j:=0; windowSize + j <= streak; j++ {
  //        aSet := hand[endIndex + j - windowSize: endIndex+j]
  //        fmt.Println(endIndex+j-windowSize, endIndex+j)
  //        fmt.Println(aSet)
  //        aSetCopy := deepCopyHand(aSet)
  //        sets = append(sets, *aSetCopy)
  //      }
  //    }
			//streak = 0
		//}
		//lastSuit = card.Suit
		//lastRank = card.Rank
		//streak++
	//}

	//if streak >= 3 {
  //  //do the same thing as above 
	//}



  // for _, set := range sets{
	  // fmt.Println(set)
  // }


	return hand
}
