function parseCraftAddress(addr) {
  if (!addr) return null;
  addr = addr.trim();

  // Match for selector:....
  let selector;
  const selectorMatch = addr.match(/^(.*?):.*$/);
  if (selectorMatch) {
    selector = selectorMatch[1].trim();
  }

  let property;
  const propertyMatch = addr.match(/^.*:(.*)?$/);
  if (propertyMatch) {
    property = propertyMatch[2].trimg();
  }
  if (property && property.beginWith("[")) {
    return {
      selector: selector,
      attr: attr,
    };
  } else {
    return {
      selector: selector,
      property: attr,
    };
  }
}

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
        fetch("/assets/main.wasm"),
        this.go.importObject,
      );
      this.go.run(result.instance);
      this.wasmBridge = globalThis.wasmBridge;
      this.isReady = true;
      console.info("Crafter: initialized");
      console.log(this.wasmBridge);
      this.call("init");
    } catch (err) {
      console.error(`Crafter: Error loading Go WASM module: ${err}`);
    }
  }

  start() {
    document.querySelectorAll("[craft-name]").forEach((ele) => {
      this.handle(ele);
    });
  }

  handle(ele) {
    if (this.isReady === false) {
      console.error("crafter: instance is not ready");
      return;
    }
    const method = ele.getAttribute("craft-name");
    console.log("craft-call", method);

    const handler = this.wasmBridge[method];
    if (!handler && typeof handler !== "function") {
      console.log("WASM method not found");
      return;
    }

    const f = () => {
      let input;
      let found = false;
      const inputID = ele.getAttribute("craft-input");
      if (inputID) {
        const inputEl = document.getElementById(inputID);
        if (inputEl) {
          input = inputEl.innerText;
          found = true;
        }
      }

      let result;
      switch (found) {
        case true:
          result = handler(input);
          break;
        case false:
          result = handler();
          break;
      }

      const target = ele.getAttribute("craft-target");
      const { selector, attribute } = parseTarget(target);
      let targetEl;
      switch (selector) {
        case "":
        case "this":
          targetEl = ele;
          break;
        default:
          targetEl = document.getElementById(selector);
          break;
      }

      switch (attribute) {
        case "":
          targetEl.innerText = result;
          break;
        default:
          targetEl.setAttribute(attribute, result);
          break;
      }
    };
    const trigger = ele.getAttribute("craft-trigger");
    switch (trigger) {
      case "":
        f();
        return;
      case "click":
        ele.addEventListener("click", () => {
          f();
        });
        return;
    }
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

  start() {
    const response = this.crafter.call("start");
    if (response.error) {
      console.error(response.error);
      return;
    }
    if (response.payload) {
      console.log(response);
    }
    this.handleFetchCard();
    this.handleUpdateCard();

    try {
      this.isReady = false;
      this.isReady = true;

      // Event Listeners
      flashcard.addEventListener("click", this.flipCard.bind(this));

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
    const app = document.getElementById("app");
    app.classList.add("hidden");
    globalThis.location.assign("/stats");
  }
  handleUpdateCard() {
    const wordEl = document.getElementById("word");
    const ipaEl = document.getElementById("ipa");
    const definitionEl = document.getElementById("definition");
    const exampleEl = document.getElementById("example");
    const flashcard = document.getElementById("flashcard");
    const playIPASoundEl = document.getElementById("play-ipa");
    ipaEl.textContent = this.currentCard.ipa;
    wordEl.textContent = this.currentCard.word;
    definitionEl.textContent = this.currentCard.definition;
    exampleEl.textContent = `"${this.currentCard.example}"`;
    flashcard.classList.remove("rotate-y-180");

    playIPASoundEl.addEventListener("click", (ev) => {
      // stop this ev propagating to parent object.
      ev.stopPropagation();
      // blur the focus status immediately.
      playIPASoundEl.blur();
    });
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
    const response = this.crafter.call(
      "submit",
      JSON.stringify(this.currentCard.ID),
      JSON.stringify(parseInt(rating)),
    );
    console.log(response);
  }

  nextCard() {
    this.handleFetchCard();
    this.handleUpdateCard();
  }
}
