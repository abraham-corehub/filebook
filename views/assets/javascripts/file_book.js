//console.log("Javascript for File Book Loaded!!!")
$(document).ready(function () {
 fnInit();
});

function fnInit() {
 console.log("Javascript for File Book Loaded!!!")
 $(document).on("change", "select", fnOnChangeSelect);

}

function fnAjaxOnSuccess(data, status) {
 console.log(this.target + ", Data: " + data.Name + "\nStatus: " + status)
}

function fnOnChangeSelect() {
 el = $(this)
 console.log("Clicked!, Selection is :" + el.closest('.qor-field__label').contents().html());
 
 /*
 d = el.attr('id').split("_")
 urlAjax = "/admin/branches?scopes=";
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
 */
 //$.post(urlAjax, dataAjax, fnAjaxOnSuccess(data, success));
}

function fnLog(logStr)
{
    var dt = new Date();
	
	var dTS = dt.getFullYear() + "/" + fn_num_to_z_pfxd_str(dt.getMonth()+1,2) + "/" + fn_num_to_z_pfxd_str(dt.getDate(), 2) + " " + fn_num_to_z_pfxd_str(dt.getHours(), 2) + ":" + fn_num_to_z_pfxd_str(dt.getMinutes(), 2) + ":" + fn_num_to_z_pfxd_str(dt.getSeconds(), 2) + "." + fn_num_to_z_pfxd_str(dt.getMilliseconds(), 3);
    //var date_str = Date($.now());
    console.log("@"+datetimestamp + "> " + logStr);
}
