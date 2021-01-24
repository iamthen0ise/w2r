function listenActions() {
  document.addEventListener("submit", (e) => {
    function showData(tabs) {
      const tags = document.getElementById("w2r-tags").value;
      browser.tabs.sendMessage(tabs[0].id, {
        command: "showData",
        tags: tags,
      });
    }

    /**
     * Just log the error to the console.
     */
    function reportError(error) {
      console.error(`Could not showData: ${error}`);
    }

    /**
     * Get the active tab,
     * then call "showData()" as appropriate.
     */
    browser.tabs
      .query({ active: true, currentWindow: true })
      .then(showData)
      .catch(reportError);
  });
}

function reportExecuteScriptError(error) {
  document.querySelector("#popup-content").classList.add("hidden");
  document.querySelector("#error-content").classList.remove("hidden");
  console.error(`Failed to execute showData content script: ${error.message}`);
}

browser.tabs
  .executeScript({ file: "/content_scripts/content-script.js" })
  .then(listenActions)
  .catch(reportExecuteScriptError);
