export default defineNuxtPlugin((nuxtApp) => {
  //Get logged In User
  let getUser = () => {
    let token = localStorage.getItem("token");
    if (!token) {
      //console.log("No Token");
      return false;
    }
    return nuxtApp.$axiosApi.get("/logged-in").then(
      (res) => {
        useAuthStore().setUser(res.data);
        return true;
      },
      (err) => {
        //Remove Token If Unauthorized
        if (err.response?.status == 401 || err.status == 401) {
          localStorage.removeItem("token");
          return false;
        }
        return false;
      }
    );
  };

  //Universal Auth Guard
  let checkLogin = (softCheck=false) => {
    return new Promise(async(resolve, reject) => {
      let token = localStorage.getItem("token");
      if (!token) {
        reject("Logged Out");
      }
      let loggedIn = useAuthStore().user;
      if (!loggedIn) {
        const user = await getUser();
        user ? resolve("Logged In") : reject("Logged Out");
      } else {
        resolve("logged In");
      }
    });
  };

  //Login User
  let login = (user, pass) => {
    return new Promise((resolve, reject) => {
      nuxtApp.$axiosApi
        .post("/login", {
          username: user,
          password: pass,
        })
        .then(
          async (res) => {
            //console.log("token",res.data.success.token);
            localStorage.setItem("token", res.data.token);
            getUser();
            resolve(res);
          },
          (err) => {
            reject(createError("Error Logging In: " + err));
            if(err.response?.status===401){
              useToaster('Please Recheck Your E-Mail & Password!',"bg-red-500");
            } else {
              useToaster('An Error Occured!',"bg-red-500");
            }
          }
        );
    });
  };
  let logout = ()=>{
    navigateTo("/");
    localStorage.removeItem("token");
    useAuthStore().setUser(null);
    useNuxtApp().$bus.$emit('toaster',['Logged Out!','info']);
  }
  return {
    provide: {
      login,
      getUser,
      checkLogin,
      logout
    },
  };
});
