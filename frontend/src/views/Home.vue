
<template>

 <main class="home-page">
  <div className="Dashboard-info">

    <div className="isp-info">
      <div className="ispad"><b>{{ISP_info}}</b></div>

    </div>
     <div className="public-ip">
      <div className="publicad">
       <b >PUBLIC IP {{public_ip}}</b>
      </div>


     </div>
    <div className="internet-status">
      {{LAN_network}}
    </div>
    </div>
  </main >
</template>

<script>
// import axios from 'axios';
// export default {
//   data() {
//     return {
//         public_ip: "",
//         ISP_info:"",
//         LAN_network:""
//     }
//   },
//   mounted() {
//     this.loadDash();
//   },
//   methods:{
//     loadDash:function(){
//     axios.get("http://localhost:3000/api/ip").then((resp) => {
//       this.public_ip=resp.data;
//     })
//     .catch(()=>{
//       console.log("something went wrong")
//     })
//     axios.get('http://localhost:3000/api/isp').then((resp) => {
//       this.ISP_info=resp.data;
//     })
//     .catch(()=>{
//       console.log("something went wrong")
//     })
//     axios.get('http://localhost:3000/api/lan').then((resp) => {
//         this.LAN_network=resp.data;
//     })
//     .catch(()=>{
//       console.log("something went wrong")
//     })
//   }
// }
// }
// _________________________
// import socket from 'socket.io-client';
// import axios from 'axios';

// export default {
//   data() {
//     return {
//       public_ip: "",
//       ISP_info:"",
//       LAN_network:""
//     };
//   },
//   created() {
//     this.$socket.on('public_ip', (data) => {
//       this.public_ip = data.ip;
//     });
//     axios.get('http://localhost:3000/api/ip')
//     .then(response => {
//       // this.public_ip=response.data;
//       console.log(response.data);
//     })
//     .catch(error => {
//       console.log(error);
//     });
//   },
// };

</script>

<style scoped>
.home-page {
 overflow: hidden;
  background-color:lightblue;
}
.Dashboard-info{
    height: 100vh;
    /* background-color: #28282B ; */
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: center;
    padding: 30px;
    display: flex;
    /* flex-direction:row-reverse; */
    /* background: #214b8b; */
}
.isp-info{
    /* margin-left: 150px; */
    position: relative;
    margin-left: 10px;
    height: 200px;
    width: 250px;
    background-color: #45B8AC;
    border-radius: 6px;
    box-shadow: -3px -3px 7px #fdfdfdb2,
     3px 3px 5px rgba(94, 104, 121, 0.945);
}
.public-ip{
  position: relative;
    margin-left: 25px;
    height: 200px;
    width: 250px;
    border-radius: 6px;
    box-shadow: -3px -3px 7px #ffffffb2,
     3px 3px 5px rgba(94, 104, 121, 0.945);
     background-color: #45B8AC;
}
.publicad{
  position: relative;
  padding: 80px;
  right: 20px;

}
.ispad{
  position: relative;
  padding: 40px;
  left: 12px;
}
.internet-status{

    margin: 25px;
    height:200px;
    width: 250px;
    border-radius: 6px;
    box-shadow: -3px -3px 7px #ffffffb2,
     3px 3px 5px rgba(94, 104, 121, 0.945);
     background-color: #45B8AC;
}
</style>
