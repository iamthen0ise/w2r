(function () {
  if (window.hasRun) {
    return;
  }
  window.hasRun = true;

  let token = "";
  browser.storage.sync.get("w2rDataNeed").then((tk) => {
    token = tk["w2rDataNeed"];
  });

  function showData(message) {
    const tags = message.tags || "unsorted";
    fetch("https://api.github.com/repos/iamthen0ise/w2r/dispatches", {
      method: "POST",
      headers: {
        Accept: "application/vnd.github.v3+json",
        Authorization: token,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        event_type: "webhook",
        client_payload: {
          url: document.URL,
          title: document.title,
          tags: tags,
        },
      }),
    })
      .then((response) => console.log(response))
      .catch((error) => console.error(error));
  }

  browser.runtime.onMessage.addListener((message) => {
    if (message.command === "showData") {
      showData(message);
    }
  });
})();
