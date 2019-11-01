//alert("Javascript for File Book Loaded!!!")
$(document).ready(function () {
 console.log("Javascript for File Book Loaded!!!")
 $(document).on("change", "select", function () {
  console.log("Clicked!, Selection is :" + this.id);
  d = this.id.split("_")
  /*$.ajax({
   type: "POST",
   url: "admin/ajax",
   data: data,
   success: success,
   dataType: dataType
 });
 */
  $.post("/admin/ajax",
   {
    res: d[0],
    id: d[1],
    field: d[2],
    value: this.value
   },
   function (data, status) {
    console.log("Data: " + data.Name + "\nStatus: " + status);
   });
 });
 /*
 $('select').on('change', function() {
  alert( this.value );
 });
*/
});