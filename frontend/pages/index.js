import { useEffect } from "react";
import { ethers } from "ethers";
import useEthereum from "@/hooks/useethereum";

export default function Home() {
  const { balance, getBalance, address, provider, connect } = useEthereum();

  useEffect(() => {
    connect()
  }, [])

  useEffect(() => {
    if (address) {
      getBalance();
    }
  }, [address])

  return (
    <div>
      <Navigation />
      <div className="container mx-auto">
        <div className="flex space-x-3">
          <div className="w-1/3 p-10 bg-teal-400 rounded-xl">
            balance: {balance.toString()}
          </div>
          <PlaceOrderCard />
        </div>
      </div>
    </div>
  )
}

const PlaceOrderCard = () => {
  return (
    <div className="w-1/3 p-6 bg-teal-400 rounded-xl text-white">
      <div>
        <div>Place market order</div>
        <div className="py-2"></div>
      </div>
    </div>
  )
}

const Button = ({ children, onClk }) => {
  return (
    <button onClick={onClk} className="p-3 bg-blue-500 font-bold rounded-lg text-white">
      {children}
    </button>
  )
}

const Navigation = () => {
  const { connect, address } = useEthereum();
  return (
    <div className="container mx-auto py-8 mb-20">
      <div className="flex justify-between">
        <div>
          <a href="#">ExchangeTT</a>
        </div>
        <div className="flex space-x-6">
          <a href="#">portfolio</a>
          <a href="#">help</a>
          {address ? (
            <div>{address}</div>
          ): (
            <Button onClk={() => connect()}>Connect</Button>  
          )}
        </div>
      </div>
    </div>
  )
}


