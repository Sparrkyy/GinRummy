package main

import "testing"
//import "fmt"

func TestMakeDeck(t *testing.T) {
  var deck *[]Card;
	deck = makeDeck()
	if len(*deck) != 52 {
		t.Fatal("Deck is not 52 cards")
	}

  var hand1 []Card;
	hand1, deck = MakeHand(deck)
	if len(hand1) != 10 {
		t.Fatal("hand was not 10 cards")
	}

	if len(*deck) != 42 {
		t.Fatal("Deck did not go down after creating the hand")
	}

  var hand2 []Card;
	hand2, deck = MakeHand(deck)
	if len(hand2) != 10 {
		t.Fatal("hand was not 10 cards")
	}

	if len(*deck) != 32 {
		t.Fatal("Deck did not go down after creating the hand 2")
	}
}

func TestPopCard(t *testing.T){
  var deck *[]Card;
	deck = makeDeck()
  _, *deck = PopCardStack(deck)
  if len(*deck) != 51 {
    t.Fatal("Deck did not lose a card to pop")
  }
}

func TestRemoveCard(t *testing.T){
  deck := makeDeck()
	if len(*deck) != 52 {
		t.Fatal("Deck is not 52 cards")
	}
  spades5 := Card{Rank:5, Suit:Spades}
  var newDeck []Card;
  newDeck, err := removeCard(*deck, spades5)

  if err != nil {
    t.Fatal("There was a error removing the card", err.Error())
  }

  if len(newDeck) != 51 {
    t.Fatal("The Length of the deck was not one less after trying to filter one card, it was: ", len(*deck))
  }

  for _, card := range newDeck{
    if equalCards(card, spades5){
      t.Fatal("Found the card that was supposed to be removed");
    }
  }
  
}

func TestFindHandScore(t *testing.T){
  // deck := makeDeck()
  // hand,deck := MakeHand(deck);
  // hand = findHandScore(hand)
  hand := []Card{
    {1, Spades},
    {1, Hearts},
    {1, Clubs},
    {1, Diamonds},
    {2, Clubs},
    {3, Clubs},
    {4, Clubs},
    {5, Clubs},
    {6, Clubs},
    {9, Clubs},
  }

  hand = findHandScore(hand)

}
