package llm

import "testing"

func TestSentences(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"Una sola frase.", 1},
		{"Acumula 2.825 kWh por encima de lo esperado.", 1}, // the dot in 2.825 is not a sentence end
		{"Subió 110,5%. Cayó el FP. Sin evento.", 3},
		{"¿Y ahora? Sí!", 2},
		{"Sin punto final", 0},
		{"", 0},
	}
	for _, tt := range tests {
		if got := sentences(tt.in); got != tt.want {
			t.Errorf("sentences(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestNumberGrounding(t *testing.T) {
	set, err := numbersIn(sampleResult())
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		text string
		ok   bool
	}{
		{"110,5", true}, {"110.5", true}, {"111", true}, // rounding
		{"2.825", true}, {"2,825", true}, {"2825", true}, // thousands in either convention
		{"1.047,7", true}, {"1,047.7", true},
		{"0,94", true}, {"94", true}, {"0,74", true}, // and a fraction written as a percentage
		{"-21,3", true}, {"222,6", true}, // signs are ignored
		{"58", true}, {"35", true}, // hours and z from the evidence
		{"2026", true}, {"12", true}, {"14", true}, {"9", true}, // parts of the timestamps
		{"5", true}, {"31", true}, // small whole numbers are always allowed
		{"37", false}, {"1.234,5", false}, {"173,2", false}, {"987.654", false}, {"2027", false},
		{"0,55", false}, {"300", false},
	}
	for _, tt := range tests {
		nums := extractNumbers(tt.text)
		if len(nums) != 1 {
			t.Fatalf("%q parsed into %d numbers", tt.text, len(nums))
		}
		if got := set.has(nums[0]); got != tt.ok {
			t.Errorf("has(%q) = %v, want %v", tt.text, got, tt.ok)
		}
	}
}

func TestExtractNumbers(t *testing.T) {
	got := extractNumbers("Subió +110,5% (de 1.047,7 a 2.207,6 kWh) el 12 sep, 14:00; FP 0.74.")
	var texts []string
	for _, n := range got {
		texts = append(texts, n.text)
	}
	want := []string{"110,5", "1.047,7", "2.207,6", "12", "14", "00", "0.74"}
	if len(texts) != len(want) {
		t.Fatalf("got %v, want %v", texts, want)
	}
	for i := range want {
		if texts[i] != want[i] {
			t.Errorf("got %v, want %v", texts, want)
			break
		}
	}
}

func TestAcceptedTextEndsWithAPeriod(t *testing.T) {
	set, _ := numbersIn(sampleResult())
	rw, err := parseAndValidate(`{"reason":"Subió 110,5% el 12 sep","explanation":"Cayó el FP a 0,74. Sin evento","recommended_action":"Revisar el medidor"}`, set)
	if err != nil {
		t.Fatal(err)
	}
	if rw.Reason != "Subió 110,5% el 12 sep." || rw.Explanation != "Cayó el FP a 0,74. Sin evento." || rw.RecommendedAction != "Revisar el medidor." {
		t.Errorf("got %+v", rw)
	}
	if _, err := parseAndValidate(`{"reason":"","explanation":"a 12","recommended_action":"b 12"}`, set); err == nil {
		t.Error("an empty field must still be rejected")
	}
}
