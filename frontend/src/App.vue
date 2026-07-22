<template>
  <div id="app">
    <a v-if="!isLoginPage" class="skip-link" href="#main-content">跳到主要内容</a>
    <div class="app-ambience" aria-hidden="true">
      <div class="bg-grid"></div>
      <div class="ambient-shape ambient-shape-one"></div>
      <div class="ambient-shape ambient-shape-two"></div>
    </div>

    <Navbar v-if="!isLoginPage" />

    <main
      id="main-content"
      :class="isLoginPage ? 'login-main' : 'app-main'"
    >
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import Navbar from "@/components/NavBar.vue";

const route = useRoute();

const isLoginPage = computed(() => route.path === "/login");

onMounted(() => {
  fetch("/api/settings/login")
    .then((response) => response.json())
    .then((result) => {
      if (result.code === 200 && result.data?.site_logo) {
        let link = document.querySelector("link[rel*='icon']");
        if (!link) {
          link = document.createElement("link");
          link.rel = "icon";
          document.head.appendChild(link);
        }
        link.href = result.data.site_logo;
      }
    })
    .catch((error) => console.error("Logo fetch failed:", error));
});
</script>

<style scoped>
#app {
  min-height: 100vh;
}

.app-ambience,
.app-ambience > div {
  position: fixed;
  inset: 0;
  pointer-events: none;
}

.app-ambience {
  z-index: 0;
  overflow: hidden;
}

.app-ambience .bg-grid {
  opacity: 0.65;
}

.ambient-shape {
  width: 27rem;
  height: 27rem;
  border-radius: 999px;
  filter: blur(90px);
  opacity: 0.26;
}

.ambient-shape-one {
  inset: -11rem auto auto -10rem !important;
  background: #f7c5d0;
}

.ambient-shape-two {
  inset: auto -12rem -13rem auto !important;
  background: #cedfd6;
}

.dark .ambient-shape {
  opacity: 0.1;
}

.app-main,
.login-main {
  position: relative;
  z-index: 1;
}

.app-main {
  min-height: 100vh;
  padding: 5.75rem 1rem 4rem;
}

.login-main {
  display: grid;
  min-height: 100vh;
  place-items: center;
}

.skip-link {
  position: fixed;
  z-index: 100;
  top: 0.55rem;
  left: 0.55rem;
  padding: 0.55rem 0.75rem;
  border-radius: 0.65rem;
  color: white;
  background: #b84c64;
  transform: translateY(-160%);
}

.skip-link:focus {
  transform: translateY(0);
}

@media (max-width: 640px) {
  .app-main {
    padding: 5.15rem 0.65rem 3rem;
  }

  .ambient-shape {
    display: none;
  }
}
</style>
