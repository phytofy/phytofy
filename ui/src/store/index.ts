// Copyright (c) 2020 OSRAM; Licensed under the MIT license.

import Vue from "vue";
import Vuex from "vuex";
import VuexPersistence from "vuex-persist";

Vue.use(Vuex);

export interface Schedule {
  id: string;
  selected: boolean;
  startDate: string;
  stopDate: string;
  startTime: string;
  stopTime: string;
  levels: number[];
  serial: number | null;
}

interface RootState {
  eulaVisible: boolean;
  eulaAccepted: boolean;
  serials: number[];
  schedules: Schedule[];
}

const persistence = new VuexPersistence<RootState>({
  storage: window.localStorage
});

const eulaAcceptedInitialState = window.localStorage.getItem("eulaAccepted") === null ? false : window.localStorage.getItem("eulaAccepted") === "true";

export default new Vuex.Store({
  state: {
    eulaVisible: !eulaAcceptedInitialState,
    eulaAccepted: eulaAcceptedInitialState,
    serials: [] as number[],
    schedules: [] as Schedule[],
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
    },
    setSchedules(state, schedules: Schedule[]) {
      state.schedules = schedules;
    }
  },
  actions: {},
  modules: {},
  plugins: [persistence.plugin]
});
