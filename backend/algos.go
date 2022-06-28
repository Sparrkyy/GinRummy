package main

import "sort"
import "fmt"
import "encoding/json"
import "math"

func deepCopyHand(hand []Card) *[]Card {
	newHand := make([]Card, len(hand))
	for i, card := range hand {
		newCard := Card{Suit: card.Suit, Rank: card.Rank}
		newHand[i] = newCard
	}
	return &newHand
}

func deepCopySets(sets [][]Card) [][]Card {
  newSets := *new([][]Card)
  for _, set := range sets{
    newSets = append(newSets, *deepCopyHand(set))
  }
  return newSets
}
 
func isValidRun(cards []Card) bool {
	for i := 1; i < len(cards); i++ {
		if cards[i].Rank != cards[i-1].Rank+1 || cards[i].Suit != cards[i-1].Suit {
			return false
		}
	}
	return true
}

func isAllSameRank(cards []Card) bool {
	theRank := cards[0].Rank
	for _, card := range cards {
		if card.Rank != theRank {
			return false
		}
	}
	return true
}

func flattenSets(sets [][]Card) []Card {
  var flattened []Card;
  for _,set := range sets{
    flattened = append(flattened, set...)
  }
  return flattened
}

func stringifyCard(card Card) string{
  returnval, _:=  json.Marshal(card)
  return string(returnval)
}

func doSetsContainDups(set1 []Card, set2 []Card) bool{
  set1map := make(map[string]bool)
  for _, card := range set1{
    set1map[stringifyCard(card)] = true
  }
  for _, card := range set2 {
    _, exists := set1map[stringifyCard(card)]
    if exists { 
      return true
    }
  }
  return false
}

func getExcludedCards(hand []Card, compare []Card) []Card{
  compareMap := make(map[string]bool)
  var excludedCards []Card;
  for _, card := range compare{
    compareMap[stringifyCard(card)] = true
  }
  for _, card := range hand {
    _, exists := compareMap[stringifyCard(card)]
    if !exists {
      excludedCards = append(excludedCards, card)
    }
  }
  return excludedCards
}

func getCardScore(card Card) int {
  if card.Rank >= 10 {
    return 10
  }
  return int(card.Rank)
}


func getPermutations(sets [][]Card, index int, curPerm [][]Card, callback func(cards [][]Card)) {
  if len(flattenSets(curPerm)) > 10 {
    return
  }
  if index >= len(sets) {
    callback(curPerm)
    return
  }
  notPicked := deepCopySets(curPerm);
  getPermutations(sets, index + 1, notPicked, callback)
  if doSetsContainDups(flattenSets(curPerm), sets[index]) {
    return
  }
  picked := deepCopySets(curPerm);
  picked = append(picked, sets[index])
  getPermutations(sets, index + 1, picked, callback)
}

func findHandScore(hand []Card) []Card {
	var sets [][]Card
	//first sorting hand by ca
	sort.Slice(hand, func(i, j int) bool {
		return hand[i].Rank < hand[j].Rank
	})
	//now finding sets within those cards if there is three
	for windowSize := 3; windowSize < len(hand); windowSize++ {
		for j := 0; windowSize+j < len(hand); j++ {
			aSet := hand[j : j+windowSize]
			if isAllSameRank(aSet) {
				aSetCopy := deepCopyHand(aSet)
				sets = append(sets, *aSetCopy)
			}
		}
	}

	//second sorting to find if there is runs
	sort.Slice(hand, func(i, j int) bool {
		if hand[i].Suit == hand[j].Suit {
			return hand[i].Rank < hand[j].Rank
		}
		return hand[i].Suit < hand[j].Suit
	})

	fmt.Println(hand)

	for windowSize := 3; windowSize < len(hand); windowSize++ {
		for j := 0; windowSize+j < len(hand); j++ {
			aSet := hand[j : j+windowSize]
			if isValidRun(aSet) {
				aSetCopy := deepCopyHand(aSet)
				sets = append(sets, *aSetCopy)
			}
		}
	}

  var emptyPerm [][]Card;
  bestScore := math.MaxInt
  var bestSet [][]Card
  getPermutations(sets, 0, emptyPerm, func (cards [][]Card) {
    exclusedCards := getExcludedCards(hand, flattenSets(cards))
    score := 0
    for _,card := range exclusedCards {
      score += getCardScore(card)
    }
    if score < bestScore {
      bestScore = score
      bestSet = cards
    }
  })
  fmt.Println(bestScore, bestSet)


	// for _, set := range sets {
	// 	fmt.Println(set)
	// }

	return hand
}
