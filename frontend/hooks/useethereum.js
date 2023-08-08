import { useState } from "react";
import { ethers }  from "ethers";

const useEthereum = () => {
    const [address, setAddress] = useState();
    const [signer, setSigner] = useState();
    const [provider, setProvider] = useState(null);
    const [balance, setBalance] = useState(0);

    const getBalance = async () => {
        const balance = await provider.getBalance(address);
        setBalance(balance);
        console.log("balance:", balance)
        return balance;
    }

    const connect = async () => {
        if (window.ethereum) {
            const provider = new ethers.BrowserProvider(window.ethereum);
            setProvider(provider);
            // It will prompt user for account connections if it isnt connected
            const signer = await provider.getSigner();
            setSigner(signer);
            // setAddress(await signer.getAddress())
            const adsr = await signer.getAddress();
            setAddress(adsr);
            await getBalance();
        }
    }

    return {
        connect,
        balance,
        address,
        getBalance,
        provider
    }
};

export default useEthereum;