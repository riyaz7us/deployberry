export const useImageReader = (input,type)=>{
	  if (input.target.files && input.target.files[0]) {
      var reader = new FileReader();
      let img;
      reader.onload = function (e) {
        img = e.target.result;
        return img;
      };
      const r = reader.readAsDataURL(input.target.files[0]);
    }
}