function saveOptions(e) {
    e.preventDefault();
    browser.storage.sync.set({
      w2rDataNeed: document.querySelector("#data-need").value
    });
  }
  
  function restoreOptions() {
  
    function setCurrentChoice(result) {
      document.querySelector("#data-need").value = result.w2rDataNeed || "";
    }
  
    function onError(error) {
      console.log(`Error: ${error}`);
    }
  
    let getting = browser.storage.sync.get("w2rDataNeed");
    getting.then(setCurrentChoice, onError);
  }
  
  document.addEventListener("DOMContentLoaded", restoreOptions);
  document.querySelector("form").addEventListener("submit", saveOptions);
  