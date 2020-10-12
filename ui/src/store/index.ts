// Copyright (c) 2020 OSRAM; Licensed under the MIT license.

import Vue from "vue";
import Vuex from "vuex";

Vue.use(Vuex);

const eulaAcceptedInitialState = window.localStorage.getItem("eulaAccepted") === null ? false : window.localStorage.getItem("eulaAccepted") === "true";

export default new Vuex.Store({
  state: {
    eulaVisible: !eulaAcceptedInitialState,
    eulaAccepted: eulaAcceptedInitialState,
    serials: [] as number[],
  },
  mutations: {
    setEulaVisible(state, eulaVisible: boolean) {
      state.eulaVisible = eulaVisible;
    },
    setEulaAccepted(state, eulaAccepted: boolean) {
      window.localStorage.setItem("eulaAccepted", eulaAccepted.toString());
      state.eulaAccepted = eulaAccepted;
    },
    setSerials(state, serials: number[]) {
      state.serials = serials;
    }
  },
  actions: {},
  modules: {}
});
