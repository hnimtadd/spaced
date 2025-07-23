class Crafter {
  constructor() {
    this.go = null;
    this.wasmBridge = null;
    this.isReady = null;
  }

  async init() {
    this.isReady = false;
    this.go = new Go();
    try {
      const result = await WebAssembly.instantiateStreaming(
        fetch("/bin/main.wasm"),
        this.go.importObject,
      );
      this.go.run(result.instance);
      this.wasmBridge = globalThis.wasmBridge;
      this.isReady = true;
      console.info("Crafter: initialized");
      console.log(this.wasmBridge);
    } catch (err) {
      console.error(`Crafter: Error loading Go WASM module: ${err}`);
    }
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
    if (!this.isReady || !this.wasmBridge) {
      const errMsg = `Bridge not ready.`;
      console.error(errMsg);
      throw new Error(errMsg);
    }
    if (!this.wasmBridge[name] && typeof this.wasmBridge[name] !== "function") {
      const errMsg = `WASM '${name}' not found or in invalid type.`;
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
  constructor(crafter) {
    this.crafter = crafter;
    this.currCard = null;
    this.isReady = false;
  }

  init() {
    try {
      this.isReady = false;
      this.isReady = true;

      // this.handlePushStateToWasm();
      this.handleFetchCard();
      const flashcard = document.getElementById("flashcard");
      const flipBtn = document.getElementById("flip-btn");
      const prevBtn = document.getElementById("prev-btn");
      const nextBtn = document.getElementById("next-btn");

      // Event Listeners
      flipBtn.addEventListener("click", this.flipCard.bind(this));
      flashcard.addEventListener("click", this.flipCard.bind(this));
      nextBtn.addEventListener("click", this.nextCard.bind(this));
      prevBtn.addEventListener("click", this.prevCard.bind(this));
      this.handleUpdateCard();
    } catch (err) {
      console.error("Error fetching or parsing JSON:", err);
    }
  }
  handleUpdateCard() {
    console.log("update");
    const wordEl = document.getElementById("word");
    const definitionEl = document.getElementById("definition");
    const exampleEl = document.getElementById("example");
    const flashcard = document.getElementById("flashcard");
    console.log(`update with card ${this.currentCard}`);
    console.log(wordEl);
    wordEl.textContent = this.currentCard.word;
    definitionEl.textContent = this.currentCard.definition;
    exampleEl.textContent = `"${this.currentCard.example}"`;
    flashcard.classList.remove("rotate-y-180");
  }

  flipCard() {
    const flashcard = document.getElementById("flashcard");
    flashcard.classList.toggle("rotate-y-180");
  }

  handleFetchCard() {
    const card = this.crafter.call("next");
    this.currentCard = JSON.parse(card);
  }

  nextCard() {
    this.handleFetchCard();
    this.handleUpdateCard();
  }

  prevCard() {
    this.handleFetchCard();
    this.handleUpdateCard();
  }
}
const crafter = new Crafter();
const worker = new Worker(crafter);

globalThis.onload = async () => {
  await crafter.init();
  worker.init();
};
