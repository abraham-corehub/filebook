//console.log("Javascript for File Book Loaded!!!")
$(document).ready(function () {
 //fnInit();
});

function fnInit() {
 console.log("Javascript for File Book Loaded!!!")
 $(document).on("change", "select", fnOnChangeSelect);

}

function fnAjaxOnSuccess(data, status) {
 console.log(this.target + ", Data: " + data.Name + "\nStatus: " + status)
}

function fnOnChangeSelect() {
 console.log("Clicked!, Selection is :" + this.id);
 d = this.id.split("_")

 urlAjax = "/admin/departments?scopes=Finance";
 dataAjax = {
  res: d[0],
  id: d[1],
  field: d[2],
  value: this.value
 };

 $.ajax({
  type: "POST",
  url: urlAjax,
  data: dataAjax,
  success: fnAjaxOnSuccess,
  dataType: "json"
 });

 //$.post(urlAjax, dataAjax, fnAjaxOnSuccess(data, success));
}
