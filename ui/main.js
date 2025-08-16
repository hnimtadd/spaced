const WASM_URL = `/assets/${globalThis.location.pathname}.wasm`;

function parseCraftAddress(addr) {
  if (!addr) return null;
  addr = addr.trim();

  const parts = addr.split(":");
  if (parts.lenth > 2 || parts.lenth <= 0) {
    throw Error("unsupported address");
  }
  const el = parts[0];
  if (parts.length == 2) {
    const property = parts[1];
    return {
      el: el,
      prop: property,
    };
  }
  return {
    el: el,
  };
}

function parseCraftInput(input) {
  // craft-input="#ipa:innerText,#ipa:[data]"
  if (!input) return [];
  input = input.trim();
  const parts = input.split(",");
  return parts.map((item) => {
    item = item.trim();
    const parsed = parseCraftAddress(item);
    const el = document.querySelector(parsed.el);
    if (!el) return "";

    if (parsed.prop === undefined) {
      return el.innerText;
    }

    if (parsed.prop in el) {
      return el[parsed.prop];
    }

    if (parsed.prop.startsWith("[") && parsed.prop.endsWith("]")) {
      return el.getAttribute(parsed.prop.slice(1, parsed.prop.length - 1));
    }
    return "";
  });
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
      if (!("instantiateStreaming" in WebAssembly)) {
        WebAssembly.instantiateStreaming = (responsePromise, importObject) => {
          responsePromise
            .then((resp) => resp.arrayBuffer())
            .then((bytes) => WebAssembly.instantiate(bytes, importObject));
        };
      }

      const result = await WebAssembly.instantiateStreaming(
        fetch(WASM_URL),
        this.go.importObject,
      );
      console.info(result);
      this.go.run(result.instance);
      this.wasmBridge = globalThis.wasmBridge;
      this.isReady = true;
      console.info("Crafter: initialized");
      this.call("init");
    } catch (err) {
      console.error(`Crafter: Error loading Go WASM module: ${err}`);
    }
  }
  start() {
    this.buildIndex();
  }
  buildIndex() {
    document.querySelectorAll("[craft-name]").forEach((ele) => {
      if (!ele.getAttribute("craft-proceed")) this.handle(ele);
    });
  }

  handle(ele) {
    if (this.isReady === false) {
      console.error("crafter: instance is not ready");
      return;
    }
    const method = ele.getAttribute("craft-name");

    const handler = this.wasmBridge[method];
    if (!handler && typeof handler !== "function") {
      console.log("WASM method not found");
      return;
    }

    const f = () => {
      let callbackFn = () => {};
      const addr = ele.getAttribute("craft-target");
      if (addr) {
        try {
          let { el, prop } = parseCraftAddress(addr);
          let targetEl;
          switch (el) {
            case "":
            case "this":
              targetEl = ele;
              break;
            default:
              targetEl = document.querySelector(el);
              break;
          }

          switch (prop) {
            case "innerText":
              callbackFn = (res) => (targetEl.innerText = res);
              break;
            case undefined:
            case null:
            case "innerHTML":
              callbackFn = (res) => (targetEl.innerHTML = res);
              break;
            default:
              if (prop.startsWith("[") && prop.endsWith("]")) {
                prop = prop.slice(1, prop.length - 1);

                callbackFn = (res) => {
                  targetEl.setAttribute(prop, res);
                };
              } else {
                callbackFn = (res) => {
                  console.log("default callback", res);
                };
              }
              break;
          }
        } catch (err) {
          console.error("failed to parse craft address", err);
          return;
        }
      }

      const input = ele.getAttribute("craft-input");
      const parsed = parseCraftInput(input);

      const isAsync = ele.getAttribute("craft-async") !== null;

      if (isAsync) {
        handler(...parsed)
          .then(callbackFn)
          .then(() => {
            ele.setAttribute("craft-proceed", true);
          })
          .then(this.buildIndex());
      } else {
        callbackFn(handler(...parsed));
        ele.setAttribute("craft-proceed", true);
        this.buildIndex();
      }
    };

    const trigger = ele.getAttribute("craft-trigger");
    switch (trigger) {
      case null:
        f();
        return;
      case "click": {
        const handler = (e) => {
          e.stopImmediatePropagation();
          f();
        };
        ele.addEventListener("click", handler);
        return;
      }
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
  handleUpdateCard() {
    const wordEl = document.getElementById("word");
    const ipaEl = document.getElementById("ipa");
    const definitionEl = document.getElementById("definition");
    const exampleEl = document.getElementById("example");
    const flashcard = document.getElementById("flashcard");
    const playIPASoundEl = document.getElementById("play-ipa");

    flashcard.classList.remove("rotate-y-180");
    ipaEl.textContent = this.currentCard.ipa;

    // hack
    ipaEl.removeAttribute("data");

    wordEl.textContent = this.currentCard.word;
    definitionEl.textContent = this.currentCard.definition;
    exampleEl.textContent = `"${this.currentCard.example}"`;
    wordEl.addEventListener("click", function (event) {
      // Check if text is selected within the child element
      const selection = globalThis.getSelection();
      const isTextSelected =
        selection.toString().length > 0 &&
        wordEl.contains(selection.anchorNode);

      if (isTextSelected) {
        event.stopPropagation(); // Prevent the event from bubbling up to the parent
      }
    });

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
      console.log("stop");
      return;
    }
    if (response.payload) {
      this.currentCard = JSON.parse(response.payload);
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
