class Crafter {
  constructor() {
    this.go = null;
    this.wasmBridge = null;
    this.isReady = null;
  }

  init() {
    this.isReady = false;
    this.go = new Go();
    WebAssembly.instantiateStreaming(
      fetch("/bin/main.wasm"),
      this.go.importObject,
    )
      .then((result) => {
        this.go.run(result.instance);
        this.wasmBridge = globalThis.wasmBridge;
        this.isReady = true;
        console.info("Crafter: initialized");
      })
      .catch((err) => {
        console.error(`Crafter: Error loading Go WASM module: ${err}`);
      });
  }

  handle(id) {
    if (this.isReady === false) {
      console.error("crafter: instance is not ready");
      return;
    }
    const ele = document.getElementById("id");
    if (!ele) {
      console.error(`ele with id: ${id} not exists.`);
    }

    const method = ele.getAttribute("craft-name");

    const handler = this.wasmBridge[method];
    if (!handler && typeof handler !== "function") {
      console.log("WASM method not found");
    }
    const result = handler();
    console.log(result);
  }

  call(name, ...args) {
    if (
      !this.isReady ||
      !this.wasmBridge ||
      (!this.wasmBridge[name] && typeof this.wasmBridge[name] !== "function")
    ) {
      const errMsg = `WASM '${name}' not found or bridge not ready.`;
      console.error(errMsg);
      throw new Error(errMsg);
    }

    try {
      const result = this.wasmBridge[name](...args);
      return result;
    } catch (err) {
      console.error(`Error calling WASM method: '${name}'`, err);
      throw err;
    }
  }
}
class Worker {
  constructor() {
    this.cards = [
      {
        word: "Ephemeral",
        definition: "Lasting for a very short time.",
        example: "The beauty of the cherry blossoms is ephemeral.",
      },
      {
        word: "Ubiquitous",
        definition: "Present, appearing, or found everywhere.",
        example: "Smartphones have become ubiquitous in modern society.",
      },
      {
        word: "Mellifluous",
        definition: "A sound that is sweet and musical; pleasant to hear.",
        example: "She had a mellifluous voice that calmed everyone.",
      },
      {
        word: "Serendipity",
        definition:
          "The occurrence of events by chance in a happy or beneficial way.",
        example:
          "Discovering the hidden cafe was a moment of pure serendipity.",
      },
      {
        word: "Petrichor",
        definition: "The pleasant, earthy smell after rain falls on dry soil.",
        example: "He loved the smell of petrichor after a summer storm.",
      },
    ];
    this.index = 0;
  }

  init() {
    const flashcard = document.getElementById("flashcard");
    const flipBtn = document.getElementById("flip-btn");
    const prevBtn = document.getElementById("prev-btn");
    const nextBtn = document.getElementById("next-btn");

    // Event Listeners
    flipBtn.addEventListener("click", this.flipCard);
    flashcard.addEventListener("click", this.flipCard);
    nextBtn.addEventListener("click", this.nextCard);
    prevBtn.addEventListener("click", this.prevCard);
    this.handleUpdateCard();
  }
  handleUpdateCard() {
    const wordEl = document.getElementById("word");
    const definitionEl = document.getElementById("definition");
    const exampleEl = document.getElementById("example");
    const card = this.cards[this.index];
    const flashcard = document.getElementById("flashcard");
    wordEl.textContent = card.word;
    definitionEl.textContent = card.definition;
    exampleEl.textContent = `"${card.example}"`;
    flashcard.classList.remove("rotate-y-180");
  }

  flipCard() {
    const flashcard = document.getElementById("flashcard");
    flashcard.classList.toggle("rotate-y-180");
  }

  nextCard() {
    this.index = (this.index + 1) % cards.length;
    this.updateCard();
  }

  prevCard() {
    this.index = (this.index - 1 + cards.length) % cards.length;
    this.updateCard();
  }
}
const crafter = new Crafter();
const worker = new Worker();

globalThis.onload = () => {
  crafter.init();
  worker.init();
};
