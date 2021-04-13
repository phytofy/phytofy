<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <div :id="id" />
</template>

<script lang="ts">
import Vue from "vue";
import Plotly from "plotly.js";

const eventsName = [
  "AfterExport",
  "AfterPlot",
  "Animated",
  "AnimatingFrame",
  "AnimationInterrupted",
  "AutoSize",
  "BeforeExport",
  "ButtonClicked",
  "Click",
  "ClickAnnotation",
  "Deselect",
  "DoubleClick",
  "Framework",
  "Hover",
  "LegendClick",
  "LegendDoubleClick",
  "Relayout",
  "Restyle",
  "Redraw",
  "Selected",
  "Selecting",
  "SliderChange",
  "SliderEnd",
  "SliderStart",
  "Transitioning",
  "TransitionInterrupted",
  "Unhover",
];

interface DirectElement {
  /* eslint-disable  @typescript-eslint/no-explicit-any */
  on: (completeName: string, listener: (...args: any[]) => void) => DirectElement;
  removeAllListeners: (completeName: string) => DirectElement;
}

const events = eventsName
  .map((event) => event.toLocaleLowerCase())
  .map((eventName) => ({
    completeName: "plotly_" + eventName,
    /* eslint-disable  @typescript-eslint/no-explicit-any */
    handler: (context: any) => (...args: any[]) => {
      context.$emit.apply(context, [eventName, ...args]);
    },
  }));

export default Vue.extend({
  name: "PhytofyPlotly",

  props: {
    data: {
      type: Array,
    },
    layout: {
      type: Object,
    },
    config: {
      type: Object,
    },
    id: {
      type: String,
      required: false,
      default: null,
    },
  },

  mounted() {
    const element = this.$el as HTMLElement;
    const data = this.data as Partial<Plotly.PlotData>[];
    Plotly.newPlot(element, data, this.layout, this.config);
    const directElement = (element as unknown) as DirectElement;
    events.forEach((event) => {
      directElement.on(event.completeName, event.handler(this));
    });
  },

  watch: {
    data: {
      handler() {
        this.react();
      },
      deep: true,
    },

    layout: {
      handler() {
        this.react();
      },
      deep: true,
    },

    config: {
      handler() {
        this.react();
      },
      deep: true,
    },
  },

  beforeDestroy() {
    const element = this.$el as HTMLElement;
    const directElement = (element as unknown) as DirectElement;
    events.forEach((event) =>
      directElement.removeAllListeners(event.completeName)
    );
    Plotly.purge(element);
  },

  methods: {
    react() {
      const element = this.$el as HTMLElement;
      const data = this.data as Partial<Plotly.PlotData>[];
      Plotly.react(element, data, this.layout, this.config);
    },
  },
});
</script>
