import { defineStore } from "pinia";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    user: null,
    dialog:false
  }),
  getters: {
    loggedIn: (state) => {
      return state.user !== null;
    },
  },
  actions: {
    setUser(e) {
      this.user = e;
    },
    logout() {
      this.user = null;
    },
    showLogin(){
      this.dialog=true;
    },
    hideLogin(){
      this.dialog=false;
    }
  },
});
