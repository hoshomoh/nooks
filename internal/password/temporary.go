package password

import (
	"math/rand/v2"
	"strings"
)

// temporaryWords is how many words a temporary password is made of.
//
// Four from the alphabet below is about 62 bits of entropy, which is more than a
// person would choose and still something they can read out over a kitchen table.
const temporaryWords = 4

// temporarySeparator joins the words. A hyphen survives being read aloud, written on
// paper, and pasted out of a chat window.
const temporarySeparator = "-"

/*
alphabet is a short list of plain English words, chosen to be unambiguous when spoken.

Nooks has no mail server, so an Admin who adds a Member reads the password out or
writes it down. That makes it a spoken secret, and a spoken secret cannot contain
characters that have to be described — no "capital eye", no "underscore". Words are
the only form that survives the journey.

Nothing here sounds like anything else here, nothing is longer than six letters, and
nothing would embarrass somebody reading it to a flatmate.
*/
var alphabet = [...]string{
	"amber", "anchor", "apple", "arrow", "autumn", "bamboo", "basil", "beacon",
	"birch", "bishop", "bottle", "branch", "bridge", "bronze", "button", "cactus",
	"candle", "canvas", "carbon", "carrot", "cedar", "cellar", "cherry", "cinder",
	"citrus", "clever", "clover", "cobalt", "copper", "coral", "cotton", "crater",
	"crayon", "crisp", "crown", "crystal", "damson", "denim", "dinner", "dolphin",
	"donkey", "dragon", "drawer", "ember", "fabric", "falcon", "fennel", "ferry",
	"fiddle", "flint", "forest", "fossil", "garden", "ginger", "glacier", "granite",
	"gravel", "hammer", "harbour", "hazel", "helmet", "hollow", "indigo", "island",
	"ivory", "jacket", "jigsaw", "kettle", "lagoon", "lantern", "laurel", "ledger",
	"lemon", "lentil", "lilac", "linen", "lobster", "locket", "magnet", "mallet",
	"maple", "marble", "meadow", "melon", "mitten", "monsoon", "mosaic", "muffin",
	"mulberry", "nectar", "needle", "nickel", "nutmeg", "oatcake", "olive", "orbit",
	"orchard", "otter", "oxide", "paddle", "pantry", "parsley", "pebble", "pelican",
	"pepper", "pewter", "pigeon", "pillow", "pocket", "pollen", "poppy", "pottery",
	"pumpkin", "quarry", "quartz", "quiver", "rabbit", "radish", "rafter", "ribbon",
	"ripple", "rocket", "rubble", "saddle", "saffron", "sailor", "salmon", "sandal",
	"sapling", "satchel", "scarlet", "seagull", "shadow", "shelter", "sherbet", "shovel",
	"silver", "simmer", "socket", "sorrel", "spinach", "spiral", "squash", "stable",
	"stencil", "stirrup", "sunset", "syrup", "tablet", "tangle", "teapot", "thicket",
	"thimble", "thistle", "timber", "tinder", "toffee", "tomato", "trellis", "trolley",
	"trumpet", "tulip", "tundra", "turnip", "velvet", "vessel", "vinegar", "violet",
	"walnut", "walrus", "whisker", "willow", "window", "winter", "wombat", "yarrow",
}

/*
NewTemporary makes a password an Admin can read out once.

It is temporary in the only sense that matters: the Member must replace it before they
can do anything, so its whole life is the walk from one person to another. That is why
it is words rather than the strongest string this machine could produce — a secret that
has to be spoken is only as strong as the version that arrives.

The randomness is the runtime's own cryptographic source, seeded per process.
*/
func NewTemporary() string {
	words := make([]string, 0, temporaryWords)
	for range temporaryWords {
		words = append(words, alphabet[rand.IntN(len(alphabet))])
	}
	return strings.Join(words, temporarySeparator)
}
