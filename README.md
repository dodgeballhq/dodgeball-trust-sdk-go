# Dodgeball Server Trust SDK for Go

[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/dodgeballhq/dodgeball-trust-sdk-go)
[![GoReportCard example](https://goreportcard.com/badge/github.com/nanomsg/mangos)](https://goreportcard.com/report/github.com/dodgeballhq/dodgeball-trust-sdk-go)

## Table of Contents

- [Purpose](#purpose)
- [Prerequisites](#prerequisites)
- [Related](#related)
- [Installation](#installation)
- [Usage](#usage)
- [API](#api)

## Purpose

[Dodgeball](https://dodgeballhq.com) enables developers to decouple security logic from their application code. This has several benefits including:

- The ability to toggle and compare security services like fraud engines, MFA, KYC, and bot prevention.
- Faster responses to new attacks. When threats evolve and new vulnerabilities are identified, your application's security logic can be updated without changing a single line of code.
- The ability to put in placeholders for future security improvements while focussing on product development.
- A way to visualize all application security logic in one place.

The Dodgeball Server Trust SDK for Go makes integration with the Dodgeball API easy and is maintained by the Dodgeball team.

## Prerequisites

You will need to obtain an API key for your application from the [Dodgeball developer center](https://app.dodgeballhq.com/developer).

## Related

Check out the [Dodgeball Trust Client SDK](https://npmjs.com/package/@dodgeball/trust-sdk-client) for how to integrate Dodgeball into your frontend applications.

## Installation

Make sure your project is using Go Modules (it will have a `go.mod` file in its
root if it already is):

```sh
go mod init
```

Then, reference dodgeball-trust-sdk-go in a Go program with `import`:

```go
import (
  "github.com/dodgeballhq/dodgeball-trust-sdk-go"
)
```

## Usage

```go
package main

import (
  "encoding/json"
  "fmt"
  "log"
  "net"
  "net/http"
  "os"
  "strings"

  "github.com/dodgeballhq/dodgeball-trust-sdk-go"
)

// Initialize the SDK with your secret API key
var dodgeballClient = dodgeball.New(os.Getenv("DODGEBALL_SECRET_KEY"), dodgeball.NewConfig())

func orders(w http.ResponseWriter, req *http.Request) {
  userIP, err := getIP(req)
  if err != nil {
    fmt.Fprintf(w, "error reading IP\n")
      return
  }
  checkpointRequest := dodgeball.CheckpointRequest{
    CheckpointName: "PLACE_ORDER",
    Event: dodgeball.CheckpointEvent{
      IP: userIP,
      Data: map[string]interface{}{
        "order": "order456",
      },
    },
    DodgeballID: req.Header.Get("x-dodgeball-id"), // Obtained from the Dodgeball Client SDK, represents the device making the request
    UserID: "user123",
    UseVerificationID: req.Header.Get("x-dodgeball-verification-id"),
  }
  // In moments of risk, call a checkpoint within Dodgeball to verify the request is allowed to proceed
  checkpointResponse, err := dodgeballClient.Checkpoint(*checkpointRequest)
  if err != nil {
    fmt.Fprintf(w, "error checking checkpoint\n")
    return
  }

  w.Header().Set("Content-Type", "application/json")
  resp := make(map[string]interface{})

  if checkpointResponse.IsAllowed() {
    // The request is allowed to proceed .. continue with order
    w.WriteHeader(http.StatusOK) // 200
    resp["order"] = "order details"
  } else if checkpointResponse.IsRunning() {
    // If the outcome is pending, send the verification to the frontend to do additional checks (such as MFA, KYC)
    w.WriteHeader(http.StatusAccepted) // 202
    resp["verification"] = checkpointResponse.Verification
  } else if checkpointResponse.IsDenied() {
    // If the request is denied, you can return the verification to the frontend to display a reason message
    w.WriteHeader(http.StatusForbidden) // 403
    resp["verification"] = checkpointResponse.Verification
  } else {
    // If the checkpoint failed, decide how you would like to proceed. You can return the error, choose to proceed, retry, or reject the request
    w.WriteHeader(http.StatusInternalServerError) // 500
    resp["message"] = checkpointResponse.Errors
  }

  jsonResp, err := json.Marshal(resp)
  if err != nil {
    log.Fatalf("error marshalling response: %s", err)
  }
  w.Write(jsonResp)
}

func main() {
  http.HandleFunc("/api/orders", orders)
  http.ListenAndServe(os.Getenv("APP_PORT"), nil)
}

// Here's a simple utility method for grabbing the originating IP address from the request.
func getIP(r *http.Request) (string, error) {
  ip := r.Header.Get("X-REAL-IP")
  netIP := net.ParseIP(ip)
  if netIP != nil {
    return ip, nil
  }

  ips := r.Header.Get("X-FORWARDED-FOR")
  splitIps := strings.Split(ips, ",")
  for _, ip := range splitIps {
    netIP := net.ParseIP(ip)
    if netIP != nil {
      return ip, nil
    }
  }

  ip, _, err := net.SplitHostPort(r.RemoteAddr)
  if err != nil {
    return "", err
  }
  netIP = net.ParseIP(ip)
  if netIP != nil {
    return ip, nil
  }
  return "", fmt.Errorf("no valid IP found")
}

```

## API

### Configuration

---

The package requires a secret API key as the first argument to the constructor.

```go
var dodgeballClient = dodgeball.New(os.Getenv("DODGEBALL_SECRET_KEY"), dodgeball.NewConfig())
```

Optionally, you can pass in several configuration options to the constructor:

```go
var dodgeballConfig = dodgeball.NewConfig()
dodgeballConfig.APIURL = "https://api.dodgeball.com"
dodgeballConfig.APIVersion = "v1"

var dodgeballClient = dodgeball.New(os.Getenv("DODGEBALL_SECRET_KEY"), dodgeballConfig)
```

| Option       | Default                       | Description                                                                                                                             |
| :----------- | :---------------------------- | :-------------------------------------------------------------------------------------------------------------------------------------- |
| `APIVersion` | `v1`                          | The Dodgeball API version to use.                                                                                                       |
| `APIURL`     | `https://api.dodgeballhq.com` | The base URL of the Dodgeball API. Useful for sending requests to different environments such as `https://api.sandbox.dodgeballhq.com`. |

### Call a Checkpoint

---

Checkpoints represent key moments of risk in an application and at the core of how Dodgeball works. A checkpoint can represent any activity deemed to be a risk. Some common examples include: login, placing an order, redeeming a coupon, posting a review, changing bank account information, making a donation, transferring funds, creating a listing.

```go
checkpointRequest := &dodgeball.CheckpointRequest{
  CheckpointName: "CHECKPOINT_NAME",
  Event: dodgeball.CheckpointEvent{
    IP: "127.0.0.1", // The IP address of the device where the request originated
    Data: map[string]interface{}{
      // Arbitrary data to send in to the checkpoint...
      "amount": 100,
      "currency": "USD",
    },
  },
  DodgeballID: req.Header.Get("x-dodgeball-id"), // Obtained from the Dodgeball Client SDK, represents the device making the request
  UserID: "user123", // When you know the ID representing the user making the request in your database (ie after registration), pass it in here. Otherwise leave it blank.
  UseVerificationID: req.Header.Get("x-dodgeball-verification-id"), // Optional, if you have a verification ID, you can pass it in here
}

checkpointResponse, err := dodgeballClient.Checkpoint(checkpointRequest)
```

| Parameter           | Required | Description                                                                                                                                                                 |
| :------------------ | :------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `CheckpointName`    | `true`   | The name of the checkpoint to call.                                                                                                                                         |
| `Event`             | `true`   | The event to send to the checkpoint.                                                                                                                                        |
| `Event.IP`          | `true`   | The IP address of the device where the request originated.                                                                                                                  |
| `Event.Data`        | `false`  | Interface containing arbitrary data to send in to the checkpoint.                                                                                                           |
| `DodgeballID`       | `true`   | A Dodgeball generated ID representing the device making the request. Obtained from the [Dodgeball Trust Client SDK](https://npmjs.com/package/@dodgeball/trust-sdk-client). |
| `UserID`            | `false`  | When you know the ID representing the user making the request in your database (ie after registration), pass it in here. Otherwise leave it blank.                          |
| `UseVerificationID` | `false`  | If a previous verification was performed on this request, pass it in here. See the [useVerification](#useverification) section below for more details.                      |

### Interpreting the Checkpoint Response

---

Calling a checkpoint creates a verification in Dodgeball. The status and outcome of a verification determine how your application should proceed. Continue to [possible checkpoint responses](#possible-checkpoint-responses) for a full explanation of the possible status and outcome combinations and how to interpret them.

```go
type CheckpointResponse struct {
  Success bool `json:"success"`
  Errors  []struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
  } `json:"errors"`
  Version      string `json:"version"`
  Verification struct {
    ID      string `json:"id"`
    Status  string `json:"status"`
    Outcome string `json:"outcome"`
  } `json:"verification"`
}
```

| Property               | Description                                                                                                                       |
| :--------------------- | :-------------------------------------------------------------------------------------------------------------------------------- |
| `Success`              | Whether the request encountered any errors was successful or failed.                                                              |
| `Errors`               | If the `success` flag is `false`, this will contain an array of error objects each with a `code` and `message`.                   |
| `Version`              | The version of the Dodgeball API that was used to make the request. Default is `v1`.                                              |
| `Verification`         | Object representing the verification that was performed when this checkpoint was called.                                          |
| `Verification.ID`      | The ID of the verification that was created.                                                                                      |
| `Verification.Status`  | The current status of the verification. See [Verification Statuses](#verification-statuses) for possible values and descriptions. |
| `Verification.Outcome` | The outcome of the verification. See [Verification Outcomes](#verification-outcomes) for possible values and descriptions.        |

#### Verification Statuses

| Status     | Description                                                      |
| :--------- | :--------------------------------------------------------------- |
| `COMPLETE` | The verification was completed successfully.                     |
| `PENDING`  | The verification is currently processing.                        |
| `BLOCKED`  | The verification is waiting for input from the user.             |
| `FAILED`   | The verification encountered an error and was unable to proceed. |

#### Verification Outcomes

| Outcome    | Description                                                                                     |
| :--------- | :---------------------------------------------------------------------------------------------- |
| `APPROVED` | The request should be allowed to proceed.                                                       |
| `DENIED`   | The request should be denied.                                                                   |
| `PENDING`  | A determination on how to proceed has not been reached yet.                                     |
| `ERROR`    | The verification encountered an error and was unable to make a determination on how to proceed. |

#### Possible Checkpoint Responses

##### Approved

```go
checkpointResponse := &dodgeball.CheckpointResponse{
  Success: true,
  Errors: nil,
  Version: "v1",
  Verification: {
    ID: "def456",
    Status: "COMPLETE",
    Outcome: "APPROVED",
  },
}
```

When a request is allowed to proceed, the verification `status` will be `COMPLETE` and `outcome` will be `APPROVED`.

##### Denied

```go
checkpointResponse := &dodgeball.CheckpointResponse{
  Success: true,
  Errors: nil,
  Version: "v1",
  Verification: {
    ID: "def456",
    Status: "COMPLETE",
    Outcome: "DENIED",
  },
}
```

When a request is denied, verification `status` will be `COMPLETE` and `outcome` will be `DENIED`.

##### Pending

```go
checkpointResponse := &dodgeball.CheckpointResponse{
  Success: true,
  Errors: nil,
  Version: "v1",
  Verification: {
    ID: "def456",
    Status: "PENDING",
    Outcome: "PENDING",
  },
}
```

If the verification is still processing, the `status` will be `PENDING` and `outcome` will be `PENDING`.

##### Blocked

```go
checkpointResponse := &dodgeball.CheckpointResponse{
  Success: true,
  Errors: nil,
  Version: "v1",
  Verification: {
    ID: "def456",
    Status: "BLOCKED",
    Outcome: "PENDING",
  },
}
```

A blocked verification requires additional input from the user before proceeding. When a request is blocked, verification `status` will be `BLOCKED` and the `outcome` will be `PENDING`.

##### Undecided

```go
checkpointResponse := &dodgeball.CheckpointResponse{
  Success: true,
  Errors: nil,
  Version: "v1",
  Verification: {
    ID: "def456",
    Status: "COMPLETE",
    Outcome: "PENDING",
  },
}
```

If the verification has finished, with no determination made on how to proceed, the verification `status` will be `COMPLETE` and the `outcome` will be `PENDING`.

##### Error

```go
checkpointResponse := &dodgeball.CheckpointResponse{
  Success: false,
  Errors: []struct {
    Code    int
    Message string
  } {
    {
      Code: 503
      Message: "[Service Name]: Service is unavailable",
    },
  },
  Version: "v1",
  Verification: {
    ID: "def456",
    Status: "FAILED",
    Outcome: "ERROR",
  },
}
```

If a verification encounters an error while processing (such as when a 3rd-party service is unavailable), the `success` flag will be false. The verification `status` will be `FAILED` and the `outcome` will be `ERROR`. The `errors` array will contain at least one object with a `code` and `message` describing the error(s) that occurred.

### Utility Methods

---

There are several utility methods available to help interpret the checkpoint response. It is strongly advised to use them rather than directly interpreting the checkpoint response.

#### `checkpointResponse.IsAllowed()`

The `IsAllowed` method takes in a checkpoint response and returns `true` if the request is allowed to proceed.

#### `checkpointResponse.IsDenied()`

The `IsDenied` method takes in a checkpoint response and returns `true` if the request is denied and should not be allowed to proceed.

#### `checkpointResponse.IsRunning()`

The `IsRunning` method takes in a checkpoint response and returns `true` if no determination has been reached on how to proceed. The verification should be returned to the frontend application to gather additional input from the user. See the [useVerification](#useverification) section for more details on use and an end-to-end example.

#### `checkpointResponse.IsUndecided()`

The `IsUndecided` method takes in a checkpoint response and returns `true` if the verification has finished and no determination has been reached on how to proceed. See [undecided](#undecided) for more details.

#### `checkpointResponse.HasError()`

The `HasError` method takes in a checkpoint response and returns `true` if it contains an error.

#### `checkpointResponse.IsTimeout()`

The `IsTimeout` method takes in a checkpoint response and returns `true` if the verification has timed out. At which point it is up to the application to decide how to proceed.

### useVerification

---

Sometimes additional input is required from the user before making a determination about how to proceed. For example, if a user should be required to perform 2FA before being allowed to proceed, the checkpoint response will contain a verification with `status` of `BLOCKED` and outcome of `PENDING`. In this scenario, you will want to return the verification to your frontend application. Inside your frontend application, you can pass the returned verification directly to the `dodgeball.handleVerification()` method to automatically handle gathering additional input from the user. Continuing with our 2FA example, the user would be prompted to select a phone number and enter a code sent to that number. Once the additional input is received, the frontend application should simply send along the ID of the verification performed to your API. Passing that verification ID to the `useVerification` option will allow that verification to be used for this checkpoint instead of creating a new one. This prevents duplicate verifications being performed on the user.

**Important Note:** To prevent replay attacks, each verification ID can only be passed to `useVerification` once.

#### End-to-End Example

```js
// In your frontend application...
const placeOrder = async (order, previousVerification = null) => {
  const dodgeballId = await dodgeball.getIdentity();

  const endpointResponse = await axios.post(
    "/api/orders",
    { order },
    {
      headers: {
        "x-dodgeball-id": dodgeballId,
        "x-dodgeball-verification-id": previousVerificationId,
      },
    }
  );

  dodgeball.handleVerification(endpointResponse.data.verification, {
    onVerified: async (verification) => {
      // If a verification was performed and it is approved, pass it in to your API call
      await placeOrder(order, verification);
    },
    onApproved: async () => {
      // If no additional verification was required, update the view to show that the order was placed
      console.log("Order placed!");
    },
    onDenied: async (verification) => {
      // If the action was denied, update the view to show the rejection...
      console.log("Order denied.");
    },
    onError: async (error) => {
      // If there was an error performing the verification, handle it here...
      console.log("Verification error:", error);
    },
  });
};
```
