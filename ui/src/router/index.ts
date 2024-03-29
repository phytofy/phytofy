// Copyright (c) 2020 OSRAM; Licensed under the MIT license.

import Vue from "vue";
import VueRouter, { RouteConfig } from "vue-router";

Vue.use(VueRouter);

const routes: RouteConfig[] = [
  {
    path: "/schedules",
    name: "Schedules",
    // route level code-splitting
    // this generates a separate chunk (schedules.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "schedules" */ "@/views/PhytofySchedules.vue")
  },
  {
    path: "/tools",
    name: "Tools",
    // route level code-splitting
    // this generates a separate chunk (tools.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "tools" */ "@/views/PhytofyTools.vue")
  },
  {
    path: "/information",
    name: "Information",
    // route level code-splitting
    // this generates a separate chunk (information.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "information" */ "@/views/PhytofyInformation.vue")
  }
];

const router = new VueRouter({
  mode: "abstract",
  routes
});

router.push("/schedules");

export default router;
