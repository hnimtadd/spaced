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
    const startBtn = document.getElementById("start-btn");
    startBtn.addEventListener("click", this.start.bind(this));
  }

  start() {
    const startScreen = document.getElementById("start-screen");
    const app = document.getElementById("app");

    startScreen.classList.add("hidden");
    app.classList.remove("hidden");

    const response = this.crafter.call("start");
    if (response.error) {
      console.error(response.error);
      return;
    }
    this.currentCard = JSON.parse(response.payload);
    this.handleUpdateCard();

    try {
      this.isReady = false;
      this.isReady = true;

      // this.handlePushStateToWasm();
      const nextBtn = document.getElementById("next-btn");

      // Event Listeners
      flashcard.addEventListener("click", this.flipCard.bind(this));
      nextBtn.addEventListener("click", (_) => {
        this.handleSubmitReview(1);
        this.nextCard();
      });

      const ratingBtns = document.querySelectorAll(".rating-btn");
      ratingBtns.forEach((btn) => {
        btn.addEventListener("click", (e) => {
          const rating = e.target.dataset.rating;
          this.handleSubmitReview(rating);
          this.nextCard();
        });
      });

      this.handleUpdateCard();
    } catch (err) {
      console.error("Error fetching or parsing JSON:", err);
    }
  }
  stop() {
    const endScreen = document.getElementById("end-screen");
    const app = document.getElementById("app");

    endScreen.classList.remove("hidden");
    app.classList.add("hidden");
  }
  handleUpdateCard() {
    const wordEl = document.getElementById("word");
    const definitionEl = document.getElementById("definition");
    const exampleEl = document.getElementById("example");
    const flashcard = document.getElementById("flashcard");
    wordEl.textContent = this.currentCard.word;
    definitionEl.textContent = this.currentCard.definition;
    exampleEl.textContent = `"${this.currentCard.example}"`;
    flashcard.classList.remove("rotate-y-180");
  }

  handleFetchCard() {
    const response = this.crafter.call("next");
    if (response.error) {
      console.error(response.error);
      return;
    }
    if (response.stop) {
      this.stop();
      console.log("stop");
      return;
    }
    if (response.payload) {
      this.currentCard = JSON.parse(response.payload);
      console.log(this.currentCard);
    }
  }

  flipCard() {
    const flashcard = document.getElementById("flashcard");
    flashcard.classList.toggle("rotate-y-180");
  }

  handleSubmitReview(rating) {
    this.crafter.call(
      "submit",
      JSON.stringify(this.currentCard.ID),
      JSON.stringify(parseInt(rating)),
    );
  }

  nextCard() {
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
