(function () {
  function listenActions() {
    const form = document.getElementById("w2r-form");
    form.addEventListener("submit", (e) => {
      e.preventDefault();

      function showData(tabs) {
        const tags = document.getElementById("w2r-tags").value;
        browser.tabs.sendMessage(tabs[0].id, {
          command: "showData",
          tags,
        });
      }

      function reportError(error) {
        console.error(`Failed to show data: ${error.message}`);
      }

      function closePopup() {
        window.close();
      }

      browser.tabs
        .query({ active: true, currentWindow: true })
        .then(showData)
        .catch(reportError)
        .finally(closePopup);
    });
  }

  function reportExecuteScriptError(error) {
    const popupContent = document.querySelector("#popup-content");
    const errorContent = document.querySelector("#error-content");
    popupContent.classList.add("hidden");
    errorContent.classList.remove("hidden");
    console.error(`Failed to execute content script: ${error.message}`);
  }

  browser.tabs
    .executeScript({ file: "/content_scripts/content-script.js" })
    .then(listenActions)
    .catch(reportExecuteScriptError);
})();
