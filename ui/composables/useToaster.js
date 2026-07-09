export const useToaster = async (message,cl="bg-slate-800 text-white") => {
  let toast = document.createElement("div");
  toast.classList.add("toast");
  cl = cl.split(" ");
  
  toast.classList.add(cl);
  let icon = `<Icon name="mdi:check"></Icon>`;
  toast.innerHTML = icon+message;
  document.body.appendChild(toast);
  setTimeout(() => {
    toast.remove();
  }, 5000);
  return toast;
};
