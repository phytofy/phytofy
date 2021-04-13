<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <PhytofyPlotly
    :data="plotData"
    :layout="plotLayout"
    :config="plotConfig"
    @click="pick"
  />
</template>

<script lang="ts">
import Vue from "vue";
import PhytofyPlotly from "@/components/PhytofyPlotly.vue";
import { FORMATIONS, CHANNELS, SCALE_DEFAULT } from "../shared";
import * as plotly from "plotly.js";

interface Simulation {
  irradianceMaps: number[][][];
  minima: number[];
  maxima: number[];
  means: number[];
  selectedIrradianceMap: number[][];
  selectedMinimum: number;
  selectedMaximum: number;
  selectedMean: number;
  luminairesMap: number[][];
}

interface SimulatorInterface {
  irradianceSimulation: (levels: number[], orientation: number, elevation: number, countX: number, countY: number, spacing: number) => Simulation;
  irradianceSpectrum: (levels: number[]) => number[];
}

export default Vue.extend({
  name: "PhytofySimulation",

  components: {
    PhytofyPlotly,
  },

  props: {
    value: {
      type: Object,
      required: true,
    },
    settings: {
      type: Function,
    },
  },

  data: () => ({
    x: null as plotly.Datum,
    y: null as plotly.Datum,
    plotData: [
      {
        z: [[]],
        zmin: 0,
        zmax: 120,
        type: "heatmap",
        colorscale: SCALE_DEFAULT,
        hovertemplate: "x: %{x}<br />y: %{y}<br />%: %{z}<extra></extra>",
      },
      {
        fill: "tozeroy",
        fillcolor: "#00000000",
        x: [...Array(501).keys()].map((i) => i + 300),
        y: Array(501).fill(0),
        type: "scatter",
        mode: "none",
        name: "%",
        xaxis: "x2",
        yaxis: "y2",
      },
      {
        fill: "tonexty",
        fillcolor: "#FFFFFF",
        x: [290, 810],
        y: [510, 510],
        mode: "lines",
        xaxis: "x2",
        yaxis: "y2",
      },
      {
        x: [...Array(500).keys()].map((i) => i + 300),
        y: Array(600).keys(),
        z: Array(600)
          .fill(0)
          .map(() => [...Array(500).keys()].map((i) => i * 0.002)),
        zmin: 0,
        zmax: 1,
        type: "heatmap",
        colorscale: [
          // Based on: https://academo.org/demos/wavelength-to-colour-relationship/
          /* 300 nm */ ["0.00", "rgb(32,0,32)"],
          /* 380 nm */ ["0.16", "rgb(97,0,97)"],
          /* 450 nm */ ["0.30", "rgb(0,70,255)"],
          /* 520 nm */ ["0.44", "rgb(54,255,0)"],
          /* 650 nm */ ["0.70", "rgb(255,0,0)"],
          /* 730 nm */ ["0.82", "rgb(200,0,0)"],
          /* 800 nm */ ["1.00", "rgb(64,0,0)"],
        ],
        showscale: false,
        hoverinfo: "skip",
        xaxis: "x2",
        yaxis: "y2",
      },
    ],
    plotLayout: {
      title: {
        text: "",
        font: { face: "Roboto" },
      },
      grid: {
        rows: 1,
        columns: 2,
        subplots: [["x2y2", "xy"]],
        pattern: "independent",
      },
      xaxis: {
        autorange: true,
        showgrid: false,
        showticklabels: false,
        ticks: "",
        zeroline: false,
      },
      yaxis: {
        autorange: true,
        scaleanchor: "x",
        scaleratio: 1,
        showgrid: false,
        showticklabels: false,
        ticks: "",
        zeroline: false,
      },
      xaxis2: {
        fixedrange: true,
        autorange: false,
        automargin: true,
        range: [300, 800],
        showgrid: false,
        zeroline: false,
        title: {
          text: "Wavelength [nm]",
          font: { face: "Roboto" },
        },
      },
      yaxis2: {
        fixedrange: true,
        autorange: false,
        automargin: true,
        range: [0, 120],
        showgrid: false,
        zeroline: false,
        title: {
          text: "Level [%]",
          font: { face: "Roboto" },
        },
      },
      margin: {
        b: 35,
        l: 35,
        r: 35,
        t: 35,
        pad: 0,
      },
      showlegend: false,
      autosize: true,
      shapes: [] as Record<string, unknown>[],
    },
    plotConfig: {
      modeBarButtonsToRemove: [
        "zoom2d",
        "pan2d",
        "zoomIn2d",
        "zoomOut2d",
        "autoScale2d",
        "resetScale2d",
        "toggleSpikelines",
        "hoverClosestCartesian",
        "hoverCompareCartesian",
      ],
      displayModeBar: true,
      displaylogo: false,
      responsive: true,
      modeBarButtonsToAdd: [
        {
          name: "Settings",
          icon: {
            width: 1792,
            height: 1792,
            path:
              "M491 1536l91-91-235-235-91 91v107h128v128h107zm523-928q0-22-22-22-10 0-17 7l-542 542q-7 7-7 17 0 22 22 22 10 0 17-7l542-542q7-7 7-17zm-54-192l416 416-832 832h-416v-416zm683 96q0 53-37 90l-166 166-416-416 166-165q36-38 90-38 53 0 91 38l235 234q37 39 37 91z",
            transform: "matrix(1 0 0 1 0 1)",
          },
          click: () => { return; },
        },
      ],
    },
  }),

  mounted() {
    this.update();
  },

  watch: {
    value: {
      handler() {
        this.update();
      },
      deep: true,
    },
  },

  methods: {
    update() {
      this.plotConfig.modeBarButtonsToAdd[0].click = () => {
        this.$emit("settings");
      };
      const simulation = this.updateSimulation();
      this.updateScale();
      this.updateSelection(simulation);
      this.updateTitle(simulation);
      this.updateLuminairesMap(simulation);
      this.updateIrradianceSpectrum(simulation);
    },

    updateScale() {
      this.plotData[0].colorscale = this.$props.value.layout.scale;
    },

    updateSimulation() {
      return ((window as unknown) as SimulatorInterface).irradianceSimulation(
        [100.0, 100.0, 100.0, 100.0, 100.0, 100.0],
        FORMATIONS[this.$props.value.layout.formation].orientation,
        this.$props.value.layout.elevation,
        FORMATIONS[this.$props.value.layout.formation].countX,
        FORMATIONS[this.$props.value.layout.formation].countY,
        Number(this.$props.value.layout.spacing)
      );
    },

    updateSelection(simulation: Simulation) {
      const channel = this.$props.value.layout.channel;
      simulation.selectedIrradianceMap = simulation.irradianceMaps[channel];
      simulation.selectedMinimum = simulation.minima[channel];
      simulation.selectedMaximum = simulation.maxima[channel];
      simulation.selectedMean = simulation.means[channel];
      this.plotData[0].z = simulation.selectedIrradianceMap;
    },

    updateTitle(simulation: Simulation) {
      const channel = this.$props.value.layout.channel;
      const minimum = simulation.selectedMinimum.toFixed(1);
      const maximum = simulation.selectedMaximum.toFixed(1);
      const mean = simulation.selectedMean.toFixed(1);
      const titleIrradianceMap = `Map (min: ${minimum}; max: ${maximum}; mean: ${mean})`;
      const picked = this.x === null || this.y === null;
      const at = picked ? "no point selected" : `x: ${this.x}, y: ${this.y}`;
      const titleIrradianceSpectrum = `Spectrum (${at})`;
      this.plotLayout.title.text = `${CHANNELS[channel].text} Irradiance - ${titleIrradianceSpectrum} & ${titleIrradianceMap}`;
    },

    updateLuminairesMap(simulation: Simulation) {
      const shapes = [];
      if (this.$props.value.layout.luminaires) {
        for (let i = 0; i < simulation.luminairesMap.length; i++) {
          const luminaire = simulation.luminairesMap[i];
          shapes.push({
            type: "rect",
            yref: "y",
            xref: "x",
            y0: luminaire[0],
            x0: luminaire[1],
            y1: luminaire[2],
            x1: luminaire[3],
            fillcolor: "#FFFFFF",
            opacity: 0.5,
            line: {
              width: 1,
            },
          });
        }
      }
      this.plotLayout.shapes = shapes;
    },

    updateIrradianceSpectrum(simulation: Simulation) {
      const nullableX = this.x;
      const nullableY = this.y;
      if (nullableX != null && nullableY != null) {
        const x = nullableX as number;
        const y = nullableY as number;
        const levels = [];
        for (let channel = 0; channel < CHANNELS.length; channel++) {
          if (channel === this.$props.value.layout.channel) {
            levels.push(simulation.irradianceMaps[channel][y][x]);
          } else {
            levels.push(0);
          }
        }
        const profile: number[] = ((window as unknown) as SimulatorInterface).irradianceSpectrum(
          levels
        );
        profile.push(0);
        this.plotData[1].y = profile;
      }
    },

    pick(event: plotly.PlotMouseEvent) {
      if (event.points[0].data.type == "heatmap") {
        this.x = event.points[0].x;
        this.y = event.points[0].y;
      }
      this.update();
    },
  },
});
</script>
