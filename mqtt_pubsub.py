
import time
import paho.mqtt.client as paho
import os


class MQTTPubSub(object):

    def __init__(self, host="127.0.0.1"):
        #create an mqtt client
        mypid = os.getpid()
        client_uniq = "two_queue_"+str(mypid)
        self.pubsu = paho.Client(client_uniq)
        self.pubsu.connect(host)
        self.pubsu.on_message = self.on_message
        self.channels = set()
        self.received = False
        self.message  = ""

    def on_message(self, client, obj, msg):
        self.message = msg.topic+" "+msg.payload
        self.received = True

    def publish(self, channel, message):
        self.pubsu.publish(channel, message)

    def subscribe(self, channels):
        if type(channels) == list:
            for channel in channels:
                self.channels.add(channel)
                self.pubsu.subscribe(channel)
        else:
            self.channels.add(channels)
            self.pubsu.subscribe(channels)

    def unsubscribe(self, channels):
        if type(channels) == list:
            for channel in channels:
                self.channels.remove(channel)
                self.pubsu.unsubscribe(channel)
        else:
            self.channels.remove(channels)
            self.pubsu.unsubscribe(channels)

    def pubsub(self):
        return self

    def recv(self):
        while self.received == False:
            self.pubsu.loop();
        self.received = False
        return self.message

    def listen(self):
        while True:
            channel, _, data = self.recv().partition(" ")
            yield {"type": "message", "channel": channel, "data": data}


if __name__ == "__main__":
    receiver = MQTTPubSub(host="127.0.0.1")
    pubsub = receiver.pubsub()
    last = time.time()
    messages = 0
    pubsub.subscribe(["teste"])
    try:
        while True:
            pubsub.publish("teste",pubsub.recv())
            #print pubsub.recv()
            messages += 1
            now = time.time()
            if now - last > 1:
                print "%s msg/sec" % messages
                last = now
                messages = 0
    except (KeyboardInterrupt, SystemExit):
        pass
