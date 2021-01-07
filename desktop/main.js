// Copyright (c) 2020 OSRAM; Licensed under the MIT license.

const { app, BrowserWindow } = require("electron");
const path = require("path");
const { spawn } = require("child_process");

const cliPath = path.join(path.dirname(__dirname), "cli", "phytofy-cli.exe");
const cli = spawn(cliPath, ["v1-app", "55556"], { env: { PATH: process.env.PATH } });
cli.on("close", (code) => {
  app.quit();
});

function createWindow() {
  const mainWindow = new BrowserWindow({
    width: 1280,
    height: 720,
    webPreferences: {
      contextIsolation: true,
      preload: path.join(__dirname, "preload.js")
    }
  });

  mainWindow.setMenuBarVisibility(false);
  mainWindow.loadURL("http://localhost:55556/");
}

app.whenReady().then(() => {
  createWindow();

  app.on("activate", function () {
    // On macOS it is common to re-create a window in the app when the
    // dock icon is clicked and there are no other windows open.
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  })
})

// Quit when all windows are closed.
app.on("window-all-closed", function () {
  cli.kill();
  app.quit();
});
