// Copyright (c) 2020 OSRAM; Licensed under the MIT license.

import "@mdi/font/css/materialdesignicons.css";
import Vue from "vue";
import Vuetify from "vuetify/lib";


Vue.use(Vuetify);

export default new Vuetify({
  icons: {
    iconfont: "mdi"
  },
  theme: {
    themes: {
      light: {
        // #FE5000 - PMS Orange 021 C - OSRAM Orange
        primary: "#FE5000",
        // #003E51 - PMS 3035 C
        secondary: "#003E51",
        // #8DB9CA - PMS 550 C
        tertiary: "#8DB9CA",
        // #D9D9D6 - PMS Cool Gray 1 C
        quaternary: "#D9D9D6",
        // #FE5000 - PMS Orange 021 C - OSRAM Orange
        accent: "#FE5000",
        // #D9D9D6 - PMS Cool Gray 1 C
        info: "#D9D9D6",
        // Red
        error: "#F44336",
        // Green
        success: "#4CAF50",
        // Yellow
        warning: "#FFEB3B"
      }
    }
  }
});
