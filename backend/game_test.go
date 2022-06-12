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
